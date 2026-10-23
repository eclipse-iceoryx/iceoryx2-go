# iceoryx2-go

Go Language Binding for Eclipse iceoryx2™

The package `iox2` wraps the iceoryx2 C API with cgo. Progress is tracked in
[eclipse-iceoryx/iceoryx2#2104](https://github.com/eclipse-iceoryx/iceoryx2/issues/2104).

> **Status:** early development. Only creating a `Node` is available so far.

## Requirements

* Go 1.26 or newer, with cgo enabled and a C compiler
* A stable Rust toolchain and, on Linux, `libclang-dev` and `clang`, to build
  the iceoryx2 C library
* Linux (amd64, arm64) or macOS. Windows is planned.

## Build and Test

The iceoryx2 repository is a git submodule, pinned to a release.

```sh
git clone --recurse-submodules https://github.com/eclipse-iceoryx/iceoryx2-go
cd iceoryx2-go
./build.sh
```

`build.sh` builds the static C library with
`cargo build --release --package iceoryx2-ffi-c` and then builds and tests the
binding. Afterwards, use the Go tool chain directly:

```sh
source scripts/cgo-env.sh
go test -race ./...
```

`scripts/cgo-env.sh` sets `CGO_CFLAGS` and `CGO_LDFLAGS`; in your own project
set them to the include directory of `iox2/iceoryx2.h` and to the path of
`libiceoryx2_ffi_c.a` in the same way.

## Example

```go
package main

import (
	"log"

	iox2 "github.com/eclipse-iceoryx/iceoryx2-go"
)

func main() {
	node, err := iox2.NewNodeBuilder().Name("my-node").Create(iox2.ServiceTypeIpc)
	if err != nil {
		log.Fatal(err)
	}
	defer node.Close()

	log.Printf("created node %q", node.Name())
}
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security issues are handled as
described in [SECURITY.md](SECURITY.md).

## License

Apache-2.0 OR MIT, see [LICENSE-APACHE](LICENSE-APACHE),
[LICENSE-MIT](LICENSE-MIT) and [NOTICE.md](NOTICE.md).
