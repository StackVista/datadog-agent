// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

// +build docker

package docker

import (
	"context"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/util/containers"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/docker/docker/api/types"
)

// ListSwarmServices gets a list of all swarm services on the current node using the Docker APIs.
func (d *DockerUtil) ListSwarmServices(sender aggregator.Sender) ([]*containers.SwarmService, error) {
	sList, err := d.dockerSwarmServices(sender)
	if err != nil {
		return nil, fmt.Errorf("could not get docker swarm services: %s", err)
	}

	return sList, err
}

// dockerSwarmServices returns all the swarm services in the swarm cluster
func (d *DockerUtil) dockerSwarmServices(sender aggregator.Sender) ([]*containers.SwarmService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.queryTimeout)
	defer cancel()
	sList, err := d.cli.ServiceList(ctx, types.ServiceListOptions{Status: true})
	if err != nil {
		return nil, fmt.Errorf("error listing swarm services: %s", err)
	}
	ret := make([]*containers.SwarmService, 0, len(sList))
	for _, s := range sList {
		log.Infof("value of swarm service status %s", s)
		log.Infof("value of swarm service status %s", s.ServiceStatus.DesiredTasks)
		log.Infof("value of swarm service status %s", s.ServiceStatus.RunningTasks)
		tags := []string{}
		replicaCount := s.Spec.Mode.Replicated.Replicas
		service := &containers.SwarmService{
			ID:             s.ID,
			Name:           s.Spec.Name,
			ContainerImage: s.Spec.TaskTemplate.ContainerSpec.Image,
			Labels:         s.Spec.Labels,
			Replica:		replicaCount,
			Version:        s.Version,
			CreatedAt:      s.CreatedAt,
			UpdatedAt:      s.UpdatedAt,
			Spec:           s.Spec,
			PreviousSpec:   s.PreviousSpec,
			Endpoint:       s.Endpoint,
			UpdateStatus:   s.UpdateStatus,
		}
		sender.Gauge("docker.service.replicas", float64(replicaCount), "", append(tags, "serviceName:"+s.Spec.Name))
		ret = append(ret, service)
	}

	return ret, nil
}
