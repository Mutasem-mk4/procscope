# ADR-001: eBPF over ptrace

## Status
Accepted

## Context
procscope needs to observe process behavior at runtime. The two main approaches on Linux are:
1. **ptrace** — the traditional mechanism used by strace, debuggers
2. **eBPF** — modern kernel instrumentation

## Decision
Use eBPF tracepoints for runtime observation.

## Rationale

### Performance
- ptrace intercepts every syscall with SIGSTOP/SIGCONT, causing 10-100x slowdown
- eBPF runs inline in the kernel with near-zero overhead for unmatched events
- For investigation of potentially malicious binaries, minimal interference is critical

### Visibility
- eBPF can attach to kernel tracepoints, not just syscall boundaries
- Multiple probes can run simultaneously without stacking
- Process scheduling events (fork, exec, exit) are naturally observable

### Safety
- ptrace modifies the target's execution model (SIGSTOP delivery)
- eBPF programs are verified by the kernel and cannot crash it
- eBPF cleanup is automatic on process exit

### Trade-offs
- eBPF requires elevated privileges (root or CAP_BPF)
- eBPF requires kernel 5.8+ (acceptable for target distros)
- ptrace has broader kernel support
- ptrace needs no special privileges for own processes

## Consequences
- Kernel 5.8+ is a hard requirement
- Root or specific capabilities are required
- Lower performance overhead during investigation
- More natural process tree observation
