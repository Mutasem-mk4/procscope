# ADR-003: No CGO

## Status
Accepted

## Context
CGO enables calling C code from Go but adds build complexity, cross-compilation challenges, and runtime dependencies.

## Decision
Avoid CGO entirely. Set `CGO_ENABLED=0` for all builds.

## Rationale
- **Static binaries:** No dynamic library dependencies at runtime
- **Cross-compilation:** `GOOS=linux GOARCH=arm64` works without a cross-compiler toolchain
- **Package simplicity:** No need for `-dev` packages in build dependencies
- **Reproducibility:** Fewer variables in the build environment
- **Security:** Smaller attack surface (no C code in the binary)

## Trade-offs
- Cannot use libbpf or other C eBPF libraries directly
- Some Go packages require CGO (e.g., `go-sqlite3`) — we avoid such dependencies
- Pure Go alternatives may have fewer features

## Consequences
- All dependencies must be pure Go or have pure Go alternatives
- cilium/ebpf (pure Go) is used for eBPF
- golang.org/x/sys (pure Go) is used for Linux syscalls
- Binary is fully static and portable
