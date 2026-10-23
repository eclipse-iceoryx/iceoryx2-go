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

import (
	"sync"
	"unsafe"
)

// PubSubServiceBuilder collects the settings of a publish-subscribe service.
// It holds no native resources, so it needs no cleanup. Get one from
// [Node.PublishSubscribe].
type PubSubServiceBuilder struct {
	node    *Node
	name    string
	payload PayloadType
}

// PublishSubscribe starts building a publish-subscribe service with the given
// name. A service name is a path like "Sensors/Camera/Front" of ASCII
// characters. Without a call to [PubSubServiceBuilder.Payload] the service
// transmits [BytesPayload].
func (n *Node) PublishSubscribe(serviceName string) *PubSubServiceBuilder {
	return &PubSubServiceBuilder{node: n, name: serviceName, payload: BytesPayload()}
}

// Payload sets what the service transmits, see [BytesPayload] and
// [FixedSizePayload].
func (b *PubSubServiceBuilder) Payload(payload PayloadType) *PubSubServiceBuilder {
	b.payload = payload
	return b
}

// OpenOrCreate opens the service if it exists and creates it otherwise. The
// caller owns the result and has to call [PubSubService.Close] on it.
func (b *PubSubServiceBuilder) OpenOrCreate() (*PubSubService, error) {
	return b.build(func(builder C.iox2_service_builder_pub_sub_h, handle *C.iox2_port_factory_pub_sub_h) C.int {
		return C.iox2_service_builder_pub_sub_open_or_create(builder, nil, handle)
	})
}

// Open opens a service that already exists. It fails with
// [PubSubServiceErrorDoesNotExist] otherwise, and with
// [PubSubServiceErrorIncompatibleTypes] if the service transmits a different
// payload type. The caller owns the result and has to call
// [PubSubService.Close] on it.
func (b *PubSubServiceBuilder) Open() (*PubSubService, error) {
	return b.build(func(builder C.iox2_service_builder_pub_sub_h, handle *C.iox2_port_factory_pub_sub_h) C.int {
		return C.iox2_service_builder_pub_sub_open(builder, nil, handle)
	})
}

// Create creates a service that does not exist yet. It fails with
// [PubSubServiceErrorAlreadyExists] otherwise. The caller owns the result and
// has to call [PubSubService.Close] on it.
func (b *PubSubServiceBuilder) Create() (*PubSubService, error) {
	return b.build(func(builder C.iox2_service_builder_pub_sub_h, handle *C.iox2_port_factory_pub_sub_h) C.int {
		return C.iox2_service_builder_pub_sub_create(builder, nil, handle)
	})
}

func (b *PubSubServiceBuilder) build(
	finish func(C.iox2_service_builder_pub_sub_h, *C.iox2_port_factory_pub_sub_h) C.int,
) (*PubSubService, error) {
	// Everything the C API could reject is checked before its builder is
	// created, because the builder is consumed only by the call that opens or
	// creates the service.
	if err := b.payload.validate(); err != nil {
		return nil, err
	}

	cName := C.CString(b.name)
	defer C.free(unsafe.Pointer(cName))
	var nameHandle C.iox2_service_name_h
	if rc := C.iox2_service_name_new(nil, cName, C.c_size_t(len(b.name)), &nameHandle); rc != C.IOX2_OK {
		return nil, SemanticStringError(rc)
	}
	defer C.iox2_service_name_drop(nameHandle)

	cTypeName := C.CString(b.payload.typeName)
	defer C.free(unsafe.Pointer(cTypeName))

	b.node.mu.Lock()
	defer b.node.mu.Unlock()
	if b.node.handle == nil {
		return nil, ErrClosed
	}

	serviceBuilder := C.iox2_node_service_builder(&b.node.handle, nil, C.iox2_cast_service_name_ptr(nameHandle))
	pubSubBuilder := C.iox2_service_builder_pub_sub(serviceBuilder)
	if rc := C.iox2_service_builder_pub_sub_set_payload_type_details(
		&pubSubBuilder,
		b.payload.cVariant(),
		cTypeName,
		C.c_size_t(len(b.payload.typeName)),
		C.c_size_t(b.payload.size),
		C.c_size_t(b.payload.alignment),
	); rc != C.IOX2_OK {
		// Unreachable since the payload type was validated. The builder cannot
		// be released and is leaked.
		return nil, ErrInvalidPayloadType
	}

	var handle C.iox2_port_factory_pub_sub_h
	if rc := finish(pubSubBuilder, &handle); rc != C.IOX2_OK {
		return nil, PubSubServiceError(rc)
	}
	return &PubSubService{name: b.name, payload: b.payload, handle: handle}, nil
}

// PubSubService is an open publish-subscribe service. Publishers and
// subscribers are created from it. It has to be released with
// [PubSubService.Close]. A PubSubService is safe for concurrent use.
type PubSubService struct {
	// name and payload are immutable after creation.
	name    string
	payload PayloadType

	mu     sync.Mutex
	handle C.iox2_port_factory_pub_sub_h // nil once closed
}

// Name returns the name of the service.
func (s *PubSubService) Name() string {
	return s.name
}

// Payload returns the payload type of the service.
func (s *PubSubService) Payload() PayloadType {
	return s.payload
}

// Close releases the service. Publishers and subscribers that were created from
// it stay usable. Calling Close more than once is safe and has no further
// effect.
func (s *PubSubService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handle != nil {
		C.iox2_port_factory_pub_sub_drop(s.handle)
		s.handle = nil
	}
}
