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

// #include <stdlib.h>
// #include "iox2/iceoryx2.h"
import "C"

import "unsafe"

// ServiceType selects the communication mechanism a [Node] and the services
// created from it use. The zero value is [ServiceTypeIpc].
type ServiceType int

const (
	// ServiceTypeIpc communicates between processes through shared memory.
	ServiceTypeIpc ServiceType = iota
	// ServiceTypeLocal communicates between threads of a single process.
	ServiceTypeLocal
)

func (t ServiceType) String() string {
	switch t {
	case ServiceTypeIpc:
		return "ipc"
	case ServiceTypeLocal:
		return "local"
	default:
		return "invalid"
	}
}

func (t ServiceType) toC() (C.enum_iox2_service_type_e, bool) {
	switch t {
	case ServiceTypeIpc:
		return C.iox2_service_type_e_IPC, true
	case ServiceTypeLocal:
		return C.iox2_service_type_e_LOCAL, true
	default:
		return 0, false
	}
}

// NodeBuilder collects the settings of a [Node]. It holds no native
// resources, so it needs no cleanup and can be reused to create several nodes.
type NodeBuilder struct {
	name string
}

// NewNodeBuilder returns a [NodeBuilder] with the default settings, which
// create an unnamed node.
func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{}
}

// Name sets the name of the node. A name has to consist of ASCII characters
// without NUL and must be at most 128 bytes long. [NodeBuilder.Create] reports
// a [SemanticStringError] otherwise.
func (b *NodeBuilder) Name(name string) *NodeBuilder {
	b.name = name
	return b
}

// Create creates a node that communicates over the given [ServiceType]. The
// caller owns the node and has to call [Node.Close] on it.
func (b *NodeBuilder) Create(serviceType ServiceType) (*Node, error) {
	cServiceType, ok := serviceType.toC()
	if !ok {
		// The C API terminates the process on an unknown service type.
		return nil, ErrInvalidServiceType
	}

	// The node name is created before the builder since the builder handle
	// cannot be released other than by creating the node.
	cName := C.CString(b.name)
	defer C.free(unsafe.Pointer(cName))

	var nameHandle C.iox2_node_name_h
	if rc := C.iox2_node_name_new(nil, cName, C.c_size_t(len(b.name)), &nameHandle); rc != C.IOX2_OK {
		return nil, SemanticStringError(rc)
	}
	defer C.iox2_node_name_drop(nameHandle)

	builderHandle := C.iox2_node_builder_new(nil)
	C.iox2_node_builder_set_name(&builderHandle, C.iox2_cast_node_name_ptr(nameHandle))

	var nodeHandle C.iox2_node_h
	if rc := C.iox2_node_builder_create(builderHandle, nil, cServiceType, &nodeHandle); rc != C.IOX2_OK {
		return nil, NodeCreationFailure(rc)
	}

	return newNode(nodeHandle, serviceType), nil
}
