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

// SubscriberBuilder creates a [Subscriber] of a [PubSubService]. It holds no
// native resources, so it needs no cleanup.
type SubscriberBuilder struct {
	service *PubSubService
}

// SubscriberBuilder returns a builder for a subscriber of the service.
func (s *PubSubService) SubscriberBuilder() *SubscriberBuilder {
	return &SubscriberBuilder{service: s}
}

// Create creates the subscriber. The caller owns it and has to call
// [Subscriber.Close] on it.
func (b *SubscriberBuilder) Create() (*Subscriber, error) {
	s := b.service
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handle == nil {
		return nil, ErrClosed
	}

	builder := C.iox2_port_factory_pub_sub_subscriber_builder(&s.handle, nil)
	var handle C.iox2_subscriber_h
	if rc := C.iox2_port_factory_subscriber_builder_create(builder, nil, &handle); rc != C.IOX2_OK {
		return nil, SubscriberCreateError(rc)
	}
	return &Subscriber{payload: s.payload, handle: handle}, nil
}

// Subscriber receives the payloads that the publishers of a service send. It
// has to be released with [Subscriber.Close]. A Subscriber is safe for
// concurrent use.
type Subscriber struct {
	// payload is immutable after creation.
	payload PayloadType

	mu     sync.Mutex
	handle C.iox2_subscriber_h // nil once closed
}

// Receive takes the oldest payload out of the queue of the subscriber. It does
// not block: ok is false when the queue is empty. The caller owns the sample
// and has to call [Sample.Close] on it, the subscriber hands out only a limited
// number of samples at the same time.
func (s *Subscriber) Receive() (sample *Sample, ok bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handle == nil {
		return nil, false, ErrClosed
	}

	var handle C.iox2_sample_h
	if rc := C.iox2_subscriber_receive(&s.handle, nil, &handle); rc != C.IOX2_OK {
		return nil, false, ReceiveError(rc)
	}
	if handle == nil {
		return nil, false, nil
	}
	return &Sample{handle: handle}, true, nil
}

// ReceiveValue takes the oldest payload out of the queue and returns a copy of
// it, see [Subscriber.Receive]. The service has to transmit [FixedSizePayload]
// of the type T, otherwise it fails with [ErrPayloadTypeMismatch].
func ReceiveValue[T any](s *Subscriber) (value T, ok bool, err error) {
	if !s.payload.matches(reflect.TypeFor[T]()) {
		return value, false, ErrPayloadTypeMismatch
	}
	sample, ok, err := s.Receive()
	if err != nil || !ok {
		return value, false, err
	}
	defer sample.Close()

	if payload := sample.Payload(); len(payload) == s.payload.size {
		copy(valueBytes(&value), payload)
		return value, true, nil
	}
	return value, false, ErrPayloadTypeMismatch
}

// Close releases the subscriber. Calling Close more than once is safe and has no
// further effect.
func (s *Subscriber) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handle != nil {
		C.iox2_subscriber_drop(s.handle)
		s.handle = nil
	}
}

// Sample is a payload received by a [Subscriber]. It lives in shared memory and
// has to be released with [Sample.Close], also because a sample that stays open
// keeps its service and node registered. A Sample is safe for concurrent use.
type Sample struct {
	mu     sync.Mutex
	handle C.iox2_sample_h // nil once closed
}

// Payload returns the payload of the sample, or nil if the sample was closed.
// The returned slice refers to shared memory without a copy. It must neither
// be modified nor be used after [Sample.Close], that would access memory that
// iceoryx2 has handed to a publisher again. Copy the data if it has to live
// longer.
func (s *Sample) Payload() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handle == nil {
		return nil
	}

	var src unsafe.Pointer
	C.iox2_sample_payload(&s.handle, &src, nil)
	length := int(C.iox2_sample_payload_number_of_bytes(&s.handle))
	if length == 0 {
		return []byte{}
	}
	return unsafe.Slice((*byte)(src), length)
}

// Close releases the sample. Calling Close more than once is safe and has no
// further effect.
func (s *Sample) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handle != nil {
		C.iox2_sample_drop(s.handle)
		s.handle = nil
	}
}
