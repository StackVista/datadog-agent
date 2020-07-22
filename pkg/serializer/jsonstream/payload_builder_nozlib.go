// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2019-2020 Datadog, Inc.

//+build !zlib

package jsonstream

import (
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/forwarder"
	"github.com/StackVista/stackstate-agent/pkg/serializer/marshaler"
)

// OnErrItemTooBigPolicy defines the behavior when OnErrItemTooBig occurs.
type OnErrItemTooBigPolicy int

const (
	// DropItemOnErrItemTooBig when founding an ErrItemTooBig, skips the error and continue
	DropItemOnErrItemTooBig OnErrItemTooBigPolicy = iota

	// FailOnErrItemTooBig when founding an ErrItemTooBig, returns the error and stop
	FailOnErrItemTooBig
)

// PayloadBuilder is not implemented when zlib is not available.
type PayloadBuilder struct {
}

// NewPayloadBuilder is not implemented when zlib is not available.
func NewPayloadBuilder() *PayloadBuilder {
	return nil
}

// BuildWithOnErrItemTooBigPolicy is not implemented when zlib is not available.
func (b *PayloadBuilder) BuildWithOnErrItemTooBigPolicy(marshaler.StreamJSONMarshaler, OnErrItemTooBigPolicy) (forwarder.Payloads, error) {
	return nil, fmt.Errorf("not implemented")
}
