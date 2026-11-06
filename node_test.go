// Copyright (c) 2026 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Apache Software License 2.0 which is available at
// https://www.apache.org/licenses/LICENSE-2.0, or the MIT license
// which is available at https://opensource.org/licenses/MIT.
//
// SPDX-License-Identifier: Apache-2.0 OR MIT

package iox2_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	iox2 "github.com/eclipse-iceoryx/iceoryx2-go"
)

var serviceTypes = []iox2.ServiceType{iox2.ServiceTypeIpc, iox2.ServiceTypeLocal}

func TestNodeCreateAndClose(t *testing.T) {
	for _, serviceType := range serviceTypes {
		t.Run(serviceType.String(), func(t *testing.T) {
			node, err := iox2.NewNodeBuilder().Name("go-node").Create(serviceType)
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			if got := node.Name(); got != "go-node" {
				t.Errorf("Name() = %q, want %q", got, "go-node")
			}
			if got := node.ServiceType(); got != serviceType {
				t.Errorf("ServiceType() = %v, want %v", got, serviceType)
			}
			node.Close()
		})
	}
}

func TestNodeWithoutNameIsUnnamed(t *testing.T) {
	node, err := iox2.NewNodeBuilder().Create(iox2.ServiceTypeLocal)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer node.Close()

	if got := node.Name(); got != "" {
		t.Errorf("Name() = %q, want empty", got)
	}
}

func TestNodeNameWithMaximumLength(t *testing.T) {
	name := strings.Repeat("n", 128)
	node, err := iox2.NewNodeBuilder().Name(name).Create(iox2.ServiceTypeLocal)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer node.Close()

	if got := node.Name(); got != name {
		t.Errorf("Name() = %q, want %q", got, name)
	}
}

func TestNodeNameKeepsPunctuation(t *testing.T) {
	const name = "go node/1.0_x-y"
	node, err := iox2.NewNodeBuilder().Name(name).Create(iox2.ServiceTypeLocal)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer node.Close()

	if got := node.Name(); got != name {
		t.Errorf("Name() = %q, want %q", got, name)
	}
}

func TestNodeBuilderRejectsInvalidNames(t *testing.T) {
	tests := []struct {
		name string
		node string
		want iox2.SemanticStringError
	}{
		{"too long", strings.Repeat("n", 129), iox2.SemanticStringErrorExceedsMaximumLength},
		{"not ascii", "gö-node", iox2.SemanticStringErrorInvalidContent},
		{"not utf-8", "\xff\xfe", iox2.SemanticStringErrorInvalidContent},
		{"embedded nul", "go\x00node", iox2.SemanticStringErrorInvalidContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := iox2.NewNodeBuilder().Name(tt.node).Create(iox2.ServiceTypeLocal)
			if node != nil {
				node.Close()
				t.Fatal("Create returned a node for an invalid name")
			}
			var got iox2.SemanticStringError
			if !errors.As(err, &got) || got != tt.want {
				t.Fatalf("Create error = %v, want %v", err, tt.want)
			}
			if err.Error() == "" || strings.Contains(err.Error(), "unknown") {
				t.Errorf("Error() = %q, want a message from the C API", err.Error())
			}
		})
	}
}

func TestNodeBuilderRejectsInvalidServiceType(t *testing.T) {
	for _, serviceType := range []iox2.ServiceType{-1, 2} {
		node, err := iox2.NewNodeBuilder().Create(serviceType)
		if node != nil {
			node.Close()
			t.Fatalf("Create(%d) returned a node", serviceType)
		}
		if !errors.Is(err, iox2.ErrInvalidServiceType) {
			t.Errorf("Create(%d) error = %v, want %v", serviceType, err, iox2.ErrInvalidServiceType)
		}
	}
}

func TestNodeBuilderIsReusable(t *testing.T) {
	builder := iox2.NewNodeBuilder().Name("reused")
	for range 3 {
		node, err := builder.Create(iox2.ServiceTypeLocal)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if got := node.Name(); got != "reused" {
			t.Errorf("Name() = %q, want %q", got, "reused")
		}
		node.Close()
	}
}

func TestNodeCloseIsIdempotentAndConcurrencySafe(t *testing.T) {
	node, err := iox2.NewNodeBuilder().Name("closed-twice").Create(iox2.ServiceTypeLocal)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			node.Close()
		})
	}
	wg.Wait()

	if got := node.Name(); got != "closed-twice" {
		t.Errorf("Name() after Close = %q, want %q", got, "closed-twice")
	}
}

func TestNodesCreatedConcurrently(t *testing.T) {
	for _, serviceType := range serviceTypes {
		t.Run(serviceType.String(), func(t *testing.T) {
			var wg sync.WaitGroup
			for i := range 16 {
				wg.Go(func() {
					name := fmt.Sprintf("concurrent-%d", i)
					node, err := iox2.NewNodeBuilder().Name(name).Create(serviceType)
					if err != nil {
						t.Errorf("Create(%s): %v", name, err)
						return
					}
					defer node.Close()
					if got := node.Name(); got != name {
						t.Errorf("Name() = %q, want %q", got, name)
					}
				})
			}
			wg.Wait()
		})
	}
}

func TestNodeCreateAndCloseRepeatedly(t *testing.T) {
	for range 200 {
		node, err := iox2.NewNodeBuilder().Name("loop").Create(iox2.ServiceTypeIpc)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		node.Close()
	}
}
