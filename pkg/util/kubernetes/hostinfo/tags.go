// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

// +build kubelet,kubeapiserver

package hostinfo

import (
	"fmt"
	"strings"

	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/kubelet"
)

// GetTags gets the tags from the kubernetes apiserver
func GetTags() ([]string, error) {
	labelsToTags := config.Datadog.GetStringMapString("kubernetes_node_labels_as_tags")
	if len(labelsToTags) == 0 {
		// Nothing to extract
		return nil, nil
	}

	// viper lower-cases map keys from yaml, but not from envvars
	for label, value := range labelsToTags {
		delete(labelsToTags, label)
		labelsToTags[strings.ToLower(label)] = value
	}

	nodeName, err := kubelet.HostnameProvider("")
	if err != nil {
		return nil, err
	}
	client, err := apiserver.GetAPIClient()
	if err != nil {
		return nil, err
	}
	nodeLabels, err := client.NodeLabels(nodeName)
	if err != nil {
		return nil, err
	}
	return extractTags(nodeLabels, labelsToTags), nil
}

func extractTags(nodeLabels, labelsToTags map[string]string) []string {
	var tags []string

	for labelName, labelValue := range nodeLabels {
		if tagName, found := labelsToTags[strings.ToLower(labelName)]; found {
			tags = append(tags, fmt.Sprintf("%s:%s", tagName, labelValue))
		}
	}

	return tags
}
