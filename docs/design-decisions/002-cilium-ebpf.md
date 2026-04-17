# ADR-002: cilium/ebpf Library

## Status
Accepted

## Context
Several Go eBPF libraries exist:
1. **cilium/ebpf** — Pure Go, CO-RE support, well-maintained
2. **libbpf** (via CGO) — C library, canonical, requires CGO
3. **dropbox/goebpf** — Older, less maintained
4. **iovisor/gobpf** — BCC-based, requires CGO and BCC

## Decision
Use `github.com/cilium/ebpf` as the eBPF userspace library.

## Rationale

### No CGO Required
- cilium/ebpf is pure Go — no C compiler needed at build time (only for eBPF C compilation)
- Enables static binary builds and simpler cross-compilation
- Reduces build dependencies for package maintainers

### CO-RE Support
- Full CO-RE (Compile Once Run Everywhere) with BTF
- eBPF programs compiled once, portable across kernel versions
- Uses `bpf2go` for compile-time eBPF object embedding

### Well-Maintained
- Actively maintained by the Cilium/Isovalent team
- Used in production by Cilium, Tetragon, and other major projects
- Good documentation and examples

### Trade-offs
- Less feature-complete than libbpf in some edge cases
- API can change between versions (though stabilizing)
- Requires Go module dependency

## Consequences
- No CGO in the build chain
- Static binary output (easier packaging)
- Build-time eBPF compilation via `bpf2go` (needs clang on build host)
- Go module dependency on `github.com/cilium/ebpf`
