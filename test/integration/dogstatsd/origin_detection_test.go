// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package dogstatsd

import (
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/config"
)

func TestUDSOriginDetection(t *testing.T) {
	config.SetupLogger(
		"debug",
		"",
		"",
		false,
		true,
		false,
	)

	testUDSOriginDetection(t)
}
