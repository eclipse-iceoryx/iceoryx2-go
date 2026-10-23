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
	"reflect"
	"sync"
	"unsafe"
)

// PublisherBuilder creates a [Publisher] of a [PubSubService]. It holds no
// native resources, so it needs no cleanup.
type PublisherBuilder struct {
	service *PubSubService
}

// PublisherBuilder returns a builder for a publisher of the service.
func (s *PubSubService) PublisherBuilder() *PublisherBuilder {
	return &PublisherBuilder{service: s}
}

// Create creates the publisher. The caller owns it and has to call
// [Publisher.Close] on it.
func (b *PublisherBuilder) Create() (*Publisher, error) {
	s := b.service
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handle == nil {
		return nil, ErrClosed
	}

	builder := C.iox2_port_factory_pub_sub_publisher_builder(&s.handle, nil)
	if s.payload.variant == payloadDynamic {
		// A byte payload has no length known in advance. Let the data segment
		// grow instead of failing for everything longer than the default.
		C.iox2_port_factory_publisher_builder_set_allocation_strategy(&builder, C.iox2_allocation_strategy_e_POWER_OF_TWO)
	}
	var handle C.iox2_publisher_h
	if rc := C.iox2_port_factory_publisher_builder_create(builder, nil, &handle); rc != C.IOX2_OK {
		return nil, PublisherCreateError(rc)
	}
	return &Publisher{payload: s.payload, handle: handle}, nil
}

// Publisher sends payloads to the subscribers of a service. It has to be
// released with [Publisher.Close]. A Publisher is safe for concurrent use.
type Publisher struct {
	// payload is immutable after creation.
	payload PayloadType

	mu     sync.Mutex
	handle C.iox2_publisher_h // nil once closed
}

// Send copies payload into shared memory and delivers it to the subscribers. It
// returns the number of subscribers that received it. For a service with a
// [FixedSizePayload] the payload has to have exactly the size of the type,
// use [SendValue] to send a value.
func (p *Publisher) Send(payload []byte) (recipients int, err error) {
	elements, err := p.payload.numberOfElements(len(payload))
	if err != nil {
		return 0, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle == nil {
		return 0, ErrClosed
	}

	var sample C.iox2_sample_mut_h
	if rc := C.iox2_publisher_loan_slice_uninit(&p.handle, nil, &sample, C.size_t(elements)); rc != C.IOX2_OK {
		return 0, LoanError(rc)
	}

	var dst unsafe.Pointer
	C.iox2_sample_mut_payload_mut(&sample, &dst, nil)
	if int(C.iox2_sample_mut_payload_number_of_bytes(&sample)) < len(payload) {
		C.iox2_sample_mut_drop(sample)
		return 0, LoanErrorExceedsMaxLoanSize
	}
	copy(unsafe.Slice((*byte)(dst), len(payload)), payload)

	// Sending consumes the sample, also if it fails.
	var delivered C.size_t
	if rc := C.iox2_sample_mut_send(sample, &delivered); rc != C.IOX2_OK {
		return 0, SendError(rc)
	}
	return int(delivered), nil
}

// SendValue sends value to the subscribers, see [Publisher.Send]. The service
// has to transmit [FixedSizePayload] of the type T, otherwise it fails with
// [ErrPayloadTypeMismatch].
func SendValue[T any](p *Publisher, value T) (recipients int, err error) {
	if !p.payload.matches(reflect.TypeFor[T]()) {
		return 0, ErrPayloadTypeMismatch
	}
	return p.Send(valueBytes(&value))
}

// Close releases the publisher. Calling Close more than once is safe and has no
// further effect.
func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle != nil {
		C.iox2_publisher_drop(p.handle)
		p.handle = nil
	}
}
