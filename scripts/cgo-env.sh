# shellcheck shell=bash
#
# Copyright (c) 2026 Contributors to the Eclipse Foundation
#
# See the NOTICE file(s) distributed with this work for additional
# information regarding copyright ownership.
#
# This program and the accompanying materials are made available under the
# terms of the Apache Software License 2.0 which is available at
# https://www.apache.org/licenses/LICENSE-2.0, or the MIT license
# which is available at https://opensource.org/licenses/MIT.
#
# SPDX-License-Identifier: Apache-2.0 OR MIT

# Source this file to make the Go tool chain find the iceoryx2 C library that
# build.sh built from the iceoryx2 submodule:
#
#   source scripts/cgo-env.sh
#   go test -race ./...
#
# The static library is linked by its full path. Pointing at the file instead
# of using -L/-l guarantees that the static library is used even if the
# iceoryx2 build also produced a shared library next to it.
#
# ICEORYX2_TARGET_DIR overrides the cargo target directory of the submodule.

_iox2_root="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/.." && pwd)"
_iox2_target="${ICEORYX2_TARGET_DIR:-$_iox2_root/iceoryx2/target}/release"

if [[ ! -f "$_iox2_target/libiceoryx2_ffi_c.a" ]]; then
    echo "error: $_iox2_target/libiceoryx2_ffi_c.a not found, run ./build.sh --lib-only first" >&2
    return 1 2>/dev/null || exit 1
fi

export CGO_CFLAGS="-I$_iox2_target/iceoryx2-ffi-c-cbindgen/include ${CGO_CFLAGS:-}"
export CGO_LDFLAGS="$_iox2_target/libiceoryx2_ffi_c.a ${CGO_LDFLAGS:-}"

unset _iox2_root _iox2_target
