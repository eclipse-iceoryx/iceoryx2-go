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
	"errors"
	"fmt"
	"math/bits"
	"reflect"
	"unicode/utf8"
	"unsafe"
)

// maxTypeNameLength is the longest payload type name iceoryx2 accepts, see
// MAX_TYPE_NAME_LENGTH in iceoryx2/src/constants.rs.
const maxTypeNameLength = 256

// bytesTypeName is the type name of a byte payload. It is the name Rust gives
// to u8, which makes byte payloads compatible with a Rust &[u8] service.
const bytesTypeName = "u8"

// Errors reported for payloads that cannot be transmitted.
var (
	// ErrInvalidPayloadType is returned when a [PayloadType] cannot describe
	// data in shared memory, for example because the Go type contains pointers.
	ErrInvalidPayloadType = errors.New("iox2: invalid payload type")
	// ErrPayloadTypeMismatch is returned when the payload given to a
	// [Publisher] or read from a [Subscriber] has not the type of the service.
	ErrPayloadTypeMismatch = errors.New("iox2: payload type does not match the service")
)

type payloadVariant int

const (
	payloadFixedSize payloadVariant = iota
	payloadDynamic
)

// PayloadType describes what a publish-subscribe service transmits. Publishers
// and subscribers of one service have to agree on it, iceoryx2 refuses to open a
// service whose payload type differs from the one it was created with. Create
// one with [BytesPayload] or [FixedSizePayload].
type PayloadType struct {
	variant  payloadVariant
	typeName string
	size     int
	// alignment is the alignment of the payload in shared memory.
	alignment int
	// goType is set for fixed-size payloads. It lets [SendValue] and
	// [ReceiveValue] detect a mix-up between two types of the same size.
	goType reflect.Type
	err    error
}

// BytesPayload describes a payload of arbitrary length that is transmitted as
// a sequence of bytes. The type name is "u8", like the one of a Rust byte slice.
// A [Publisher] can send slices of any length, the data segment is resized when
// necessary.
func BytesPayload() PayloadType {
	return PayloadType{variant: payloadDynamic, typeName: bytesTypeName, size: 1, alignment: 1}
}

// FixedSizePayload describes a payload that is a single value of the Go type T.
// The value is copied bytewise into shared memory, so T must consist of fixed
// size numeric and boolean fields (so no int, uint or uintptr), arrays and structs of them. Types with
// pointers, strings, slices, maps, channels, functions and interfaces are
// rejected when the service is created, as are zero-size types.
//
// The type name identifies the payload type to other processes. Two services
// are only compatible if their type names, sizes and alignments are equal. For
// communication with another language, choose the name that language uses, for
// example the one of Rust's core::any::type_name, and mirror the memory layout
// of the type, for example with a #[repr(C)] struct.
func FixedSizePayload[T any](typeName string) PayloadType {
	t := reflect.TypeFor[T]()
	p := PayloadType{
		variant:   payloadFixedSize,
		typeName:  typeName,
		size:      int(t.Size()),
		alignment: t.Align(),
		goType:    t,
	}
	if !isPlainData(t) {
		p.err = fmt.Errorf("%w: %v contains pointers or has no fixed size", ErrInvalidPayloadType, t)
	}
	return p
}

// TypeName returns the name that identifies the payload type to other
// processes.
func (p PayloadType) TypeName() string {
	return p.typeName
}

// Size returns the size in bytes of the payload, or of one element if the
// payload has a dynamic length.
func (p PayloadType) Size() int {
	return p.size
}

// Alignment returns the alignment in bytes of the payload in shared memory.
func (p PayloadType) Alignment() int {
	return p.alignment
}

// validate reports every problem the C API would otherwise answer with an
// error or by terminating the process. It has to run before the C service
// builder is created, because that cannot be released without creating the
// service.
func (p PayloadType) validate() error {
	if p.err != nil {
		return p.err
	}
	switch {
	case p.typeName == "":
		return fmt.Errorf("%w: empty type name", ErrInvalidPayloadType)
	case len(p.typeName) > maxTypeNameLength:
		return fmt.Errorf("%w: type name is longer than %d bytes", ErrInvalidPayloadType, maxTypeNameLength)
	case !utf8.ValidString(p.typeName):
		return fmt.Errorf("%w: type name is not valid UTF-8", ErrInvalidPayloadType)
	case p.size <= 0:
		return fmt.Errorf("%w: zero-size type", ErrInvalidPayloadType)
	case p.alignment <= 0 || bits.OnesCount(uint(p.alignment)) != 1:
		return fmt.Errorf("%w: alignment %d is not a power of two", ErrInvalidPayloadType, p.alignment)
	}
	return nil
}

func (p PayloadType) cVariant() C.enum_iox2_type_variant_e {
	if p.variant == payloadDynamic {
		return C.iox2_type_variant_e_DYNAMIC
	}
	return C.iox2_type_variant_e_FIXED_SIZE
}

// numberOfElements returns the number of elements to loan for a payload of
// the given length in bytes.
func (p PayloadType) numberOfElements(length int) (int, error) {
	if p.variant == payloadDynamic {
		return length, nil
	}
	if length != p.size {
		return 0, fmt.Errorf("%w: payload has %d bytes, the service transmits %d", ErrPayloadTypeMismatch, length, p.size)
	}
	return 1, nil
}

func (p PayloadType) matches(goType reflect.Type) bool {
	return p.variant == payloadFixedSize && p.goType == goType
}

// isPlainData reports whether values of t can be copied bytewise to another
// process: it must not contain any reference and, to keep the layout the same
// on every platform, no int, uint or uintptr either.
func isPlainData(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	case reflect.Array:
		return isPlainData(t.Elem())
	case reflect.Struct:
		for i := range t.NumField() {
			if !isPlainData(t.Field(i).Type) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// valueBytes returns the memory of *v as a byte slice.
func valueBytes[T any](v *T) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(v)), unsafe.Sizeof(*v))
}
