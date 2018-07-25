// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

// +build !windows,!android

//go:generate go run ../../pkg/config/render_config.go agent ../../pkg/config/config_template.yaml ./dist/datadog.yaml

package main

import (
	"os"

	"github.com/StackVista/stackstate-agent/cmd/agent/app"
)

func main() {
	// Invoke the Agent
	if err := app.AgentCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
