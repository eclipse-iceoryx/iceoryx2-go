#!/usr/bin/env bash

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

# Builds the iceoryx2 C library from the submodule and then builds and tests
# the Go binding against it.
#
#   ./build.sh            build the C library, build and test the Go binding
#   ./build.sh --lib-only build only the C library

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ICEORYX2_DIR="$SCRIPT_DIR/iceoryx2"

lib_only=false
case "${1:-}" in
    "") ;;
    --lib-only) lib_only=true ;;
    *)
        echo "usage: $0 [--lib-only]" >&2
        exit 2
        ;;
esac

if [[ ! -f "$ICEORYX2_DIR/Cargo.toml" ]]; then
    echo "==> Fetching the iceoryx2 submodule"
    git -C "$SCRIPT_DIR" submodule update --init iceoryx2
fi

echo "==> Building iceoryx2-ffi-c"
(cd "$ICEORYX2_DIR" && cargo build --release --package iceoryx2-ffi-c)

if "$lib_only"; then
    exit 0
fi

# shellcheck source=scripts/cgo-env.sh
source "$SCRIPT_DIR/scripts/cgo-env.sh"

cd "$SCRIPT_DIR"

echo "==> Building the Go binding"
go build ./...

echo "==> Running the tests"
go test -race -count=1 ./...
