// +build linux_bpf,bcc

package ebpf

import (
	"fmt"
	"unsafe"

	"github.com/StackVista/stackstate-agent/pkg/ebpf/oomkill"

	bpflib "github.com/iovisor/gobpf/bcc"
)

/*
#include <string.h>
#include "c/oom-kill-kern-user.h"
*/
import "C"

type OOMKillProbe struct {
	m      *bpflib.Module
	oomMap *bpflib.Table
}

func NewOOMKillProbe() (*OOMKillProbe, error) {
	source, err := processHeaders("pkg/ebpf/c/oom-kill-kern.c")
	if err != nil {
		return nil, fmt.Errorf("Couldn’t process headers for asset “pkg/ebpf/c/oom-kill-kern.c”: %v", err)
	}

	m := bpflib.NewModule(source.String(), []string{})
	if m == nil {
		return nil, fmt.Errorf("failed to compile “oom-kill-kern.c”")
	}

	kprobe, err := m.LoadKprobe("kprobe__oom_kill_process")
	if err != nil {
		return nil, fmt.Errorf("failed to load kprobe__oom_kill_process: %s\n", err)
	}

	if err := m.AttachKprobe("oom_kill_process", kprobe, -1); err != nil {
		return nil, fmt.Errorf("failed to attach oom_kill_process: %s\n", err)
	}

	table := bpflib.NewTable(m.TableId("oomStats"), m)

	return &OOMKillProbe{
		m:      m,
		oomMap: table,
	}, nil
}

func (k *OOMKillProbe) Close() {
	k.m.Close()
}

func (k *OOMKillProbe) GetAndFlush() []oomkill.Stats {
	results := k.Get()
	k.oomMap.DeleteAll()
	return results
}

func (k *OOMKillProbe) Get() []oomkill.Stats {
	if k == nil {
		return nil
	}

	var results []oomkill.Stats

	for it := k.oomMap.Iter(); it.Next(); {
		var stat C.struct_oom_stats

		data := it.Leaf()
		C.memcpy(unsafe.Pointer(&stat), unsafe.Pointer(&data[0]), C.sizeof_struct_oom_stats)

		results = append(results, convertStats(stat))
	}
	return results
}

func convertStats(in C.struct_oom_stats) (out oomkill.Stats) {
	out.ContainerID = C.GoString(&in.cgroup_name[0])
	out.Pid = uint32(in.pid)
	out.TPid = uint32(in.tpid)
	out.FComm = C.GoString(&in.fcomm[0])
	out.TComm = C.GoString(&in.tcomm[0])
	out.Pages = uint64(in.pages)
	out.MemCgOOM = uint32(in.memcg_oom)
	return
}