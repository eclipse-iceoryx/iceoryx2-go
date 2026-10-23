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

// The location of libiceoryx2_ffi_c.a and of the iox2/iceoryx2.h header is
// deliberately not part of the source. Provide it with CGO_CFLAGS (-I) and
// CGO_LDFLAGS (path of the static library), see README.md. Only the system
// libraries the static library depends on are listed here. They are the output
// of `cargo rustc --package iceoryx2-ffi-c --lib --crate-type staticlib --
// --print native-static-libs`.

// #cgo linux LDFLAGS: -lgcc_s -lutil -lrt -lpthread -lm -ldl
// #cgo darwin LDFLAGS: -framework Security -framework CoreFoundation -liconv
import "C"
