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
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
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

	services, err := d.cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing swarm services: %s", err)
	}
	ret := make([]*containers.SwarmService, 0, len(services))
	for _, s := range services {
		tags := []string{nil}
		activeNodes, err := getActiveNodes(ctx, d.cli)
		if err != nil {
			log.Warnf("No active nodes found")
		}

		// add the serviceId filter for Tasks
		taskFilter := filters.NewArgs()
		taskFilter.Add("ServiceID", s.ID)
		// list the tasks for that service
		tasks, err := d.cli.TaskList(ctx, types.TaskListOptions{Filters: taskFilter})

		desired := uint64(0)
		running := uint64(0)
		container := swarm.ContainerStatus{}
		// Replicated services have `Spec.Mode.Replicated.Replicas`, which should give this value.
		if s.Spec.Mode.Replicated != nil {
			desired = *s.Spec.Mode.Replicated.Replicas
		}
		for _, task := range tasks {

			// TODO: this should only be needed for "global" services. Replicated
			// services have `Spec.Mode.Replicated.Replicas`, which should give this value.
			if task.DesiredState != swarm.TaskStateShutdown {
				desired++
			}
			if _, nodeActive := activeNodes[task.NodeID]; nodeActive && task.Status.State == swarm.TaskStateRunning {
				running++
			}

			container = task.Status.ContainerStatus
		}
		log.Infof("Service %s has %d desired and %d running tasks", s.Spec.Name, desired, running)

		service := &containers.SwarmService{
			ID:             s.ID,
			Name:           s.Spec.Name,
			ContainerImage: s.Spec.TaskTemplate.ContainerSpec.Image,
			Labels:         s.Spec.Labels,
			Version:        s.Version,
			CreatedAt:      s.CreatedAt,
			UpdatedAt:      s.UpdatedAt,
			Spec:           s.Spec,
			PreviousSpec:   s.PreviousSpec,
			Endpoint:       s.Endpoint,
			UpdateStatus:   s.UpdateStatus,
			Container: 		container,
		}

		sender.Gauge("docker.service.running_replicas", float64(running), "", append(tags, "serviceName:"+s.Spec.Name))
		sender.Gauge("docker.service.desired_replicas", float64(desired), "", append(tags, "serviceName:"+s.Spec.Name))

		ret = append(ret, service)
	}

	return ret, nil
}

func getActiveNodes(ctx context.Context, c client.NodeAPIClient) (map[string]struct{}, error) {
	nodes, err := c.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return nil, err
	}
	activeNodes := make(map[string]struct{})
	for _, n := range nodes {
		if n.Status.State != swarm.NodeStateDown {
			activeNodes[n.ID] = struct{}{}
		}
	}
	return activeNodes, nil
}


