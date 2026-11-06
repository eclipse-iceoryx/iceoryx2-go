# Contributing

Contributions are welcome! The binding is young, so the most useful help is:

* 🧪 **Testing** - Add unit and integration tests, also for error paths
* 📚 **Documentation** - Improve the package documentation and add examples
* 🐛 **Bug fixes** - Report and fix issues
* ✨ **New features** - Bind more of the iceoryx2 C API, see the roadmap in
  [eclipse-iceoryx/iceoryx2#2104](https://github.com/eclipse-iceoryx/iceoryx2/issues/2104)

## Eclipse Contributor Agreement

Like every Eclipse Foundation project, iceoryx2-go requires a signed
[Eclipse Contributor Agreement](https://www.eclipse.org/legal/ECA.php). The
commits of a pull request have to be authored with the e-mail address of the
Eclipse account that signed it, otherwise the ECA check fails.

## Development Workflow

1. **Fork and clone** the repository, including the submodule:
   `git clone --recurse-submodules <your fork>`
2. **Install the prerequisites**: Go 1.26 or newer, a C compiler, a stable Rust
   toolchain and, on Linux, `libclang-dev` and `clang`
3. **Build the native library, the binding and run the tests**: `./build.sh`
4. **Make your changes** and run the linter, see below
5. **Submit a pull request** with a clear description

`./build.sh --lib-only` builds only the iceoryx2 C library. To run Go commands
yourself afterwards, source the helper that points cgo at that library:

```sh
source scripts/cgo-env.sh
go test -race ./...
```

It sets `CGO_CFLAGS` to the include directory of the generated
`iox2/iceoryx2.h` and `CGO_LDFLAGS` to the static `libiceoryx2_ffi_c.a`. No
path is baked into the Go sources.

### The iceoryx2 submodule

The `iceoryx2` submodule is pinned to a release tag. The sizes of the storage
structs in the generated C header have to match the library that is linked, so
the header and the library must always come from the same build. To try a newer
iceoryx2, check out its tag in the submodule, rebuild with `./build.sh` and
commit the new submodule pointer. The nightly workflow builds against the
iceoryx2 `main` branch to detect upcoming breaking changes early.

## Code Style

* Follow standard Go conventions ([Effective Go](https://go.dev/doc/effective_go)
  and the [Code Review Comments](https://go.dev/wiki/CodeReviewComments))
* Format with `gofmt`; `golangci-lint fmt` applies the formatters of the
  project
* Lint with [golangci-lint](https://golangci-lint.run) v2 and the pinned
  version of `.github/workflows/ci.yml`: `golangci-lint run ./...` with
  `scripts/cgo-env.sh` sourced
* Document all exported identifiers
* Use the standard `testing` package, run the tests with `-race` and do not add
  third-party test dependencies
* Wrap every native handle in a Go type with a `Close` method and make `Close`
  idempotent
* Start every file with the SPDX header used by the existing files; the linter
  checks it

## Branches and Commits

* Name branches after the issue they work on: `iox2-123-short-description`
* Start commit messages with the issue number: `[#123] Add feature description`
  and follow the
  [commit message guideline](https://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html)
* Update the unreleased section of the changelog, if there is one
