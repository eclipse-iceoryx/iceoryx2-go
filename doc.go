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

// Package iox2 provides Go bindings for [Eclipse iceoryx2], a zero-copy
// inter-process communication middleware.
//
// The package wraps the iceoryx2 C API (iceoryx2-ffi-c) with cgo. The name of
// the package matches the prefix of the C API and the namespace of the C++
// API.
//
// Building requires the static iceoryx2 C library and its header. See
// README.md and build.sh for how to build them and how to point CGO_CFLAGS
// and CGO_LDFLAGS at them.
//
// Every resource that wraps a native handle, for example a [Node], has to be
// released with its Close method, since Go has no destructors.
//
// [Eclipse iceoryx2]: https://github.com/eclipse-iceoryx/iceoryx2
package iox2
