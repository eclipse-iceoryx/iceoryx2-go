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

package iox2

// #include "iox2/iceoryx2.h"
import "C"

import (
	"sync"
)

// Node is the entry point to iceoryx2 for a process or thread. It is created
// with a [NodeBuilder] and has to be released with [Node.Close]. A Node is safe
// for concurrent use.
type Node struct {
	// name and serviceType are immutable after creation, so they can be read
	// without holding mu, also after the node was closed.
	name        string
	serviceType ServiceType

	mu     sync.Mutex
	handle C.iox2_node_h // nil once closed
}

func newNode(handle C.iox2_node_h, serviceType ServiceType) *Node {
	var nameLen C.c_size_t
	namePtr := C.iox2_node_name_as_chars(C.iox2_node_name(&handle), &nameLen)

	return &Node{
		name:        C.GoStringN(namePtr, C.int(nameLen)),
		serviceType: serviceType,
		handle:      handle,
	}
}

// Name returns the name the node was created with.
func (n *Node) Name() string {
	return n.name
}

// ServiceType returns the [ServiceType] the node was created with.
func (n *Node) ServiceType() ServiceType {
	return n.serviceType
}

// Close releases the node. Calling Close more than once is safe and has no
// further effect. Releasing a node cannot fail, so Close has no error result.
func (n *Node) Close() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.handle != nil {
		C.iox2_node_drop(n.handle)
		n.handle = nil
	}
}
