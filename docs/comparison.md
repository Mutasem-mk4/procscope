# Comparison with Existing Tools

This document honestly compares procscope with similar tools in the runtime observability and process tracing space.

## Summary Table

| Aspect | procscope | Tracee | Tetragon | Inspektor Gadget | strace | sysdig |
|--------|-----------|--------|----------|------------------|--------|--------|
| **Primary use case** | Process investigation | Runtime security | K8s security observability | K8s debugging | Syscall tracing | System troubleshooting |
| **Scope** | Single process tree | System-wide | System/pod-wide | System/pod-wide | Single process | System-wide |
| **Deployment** | CLI one-shot | Daemon/agent | K8s DaemonSet | kubectl plugin | CLI one-shot | CLI/daemon |
| **Setup complexity** | Zero | Medium | High (K8s-first) | High (K8s-first) | Zero | Low |
| **eBPF-based** | Yes | Yes | Yes | Yes | No (ptrace) | Yes |
| **Process tree tracking** | Auto-follows forks | Yes | Yes | Yes | `-f` flag | Yes |
| **Evidence bundle** | ✅ Built-in | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Markdown report** | ✅ Built-in | ❌ | ❌ | ❌ | ❌ | ❌ |
| **JSONL output** | ✅ First-class | ✅ | ✅ | ✅ | ✅ (with `-e trace=...`) | ✅ |
| **Policy engine** | ❌ Not a goal | ✅ | ✅ (TracingPolicy) | ❌ | ❌ | ✅ (chisels) |
| **K8s-native** | ❌ | ✅ | ✅ | ✅ | ❌ | Partial |
| **Blocking/enforcement** | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Dependencies** | Minimal (Go, eBPF) | Heavy | Heavy (K8s CRDs) | Heavy (K8s) | libc only | Kernel module or eBPF |
| **Container awareness** | Best-effort cgroup | Full | Full | Full | None | Full |

## Detailed Comparison

### vs. Tracee (Aqua Security)

**Tracee** is a comprehensive runtime security and forensics tool. It traces system-wide events with rich policy support.

**Where Tracee is better:**
- Full system-wide coverage
- Rich signature and policy engine
- Advanced container integration (Docker, K8s)
- Event filtering DSL
- Network packet capture
- Mature project with commercial backing

**Where procscope is better:**
- Zero setup — no config files, no policy definitions
- Process-first investigation — automatic scope to one process tree
- Evidence bundle — structured output for incident response
- Markdown report — team-ready summary
- Simpler mental model — investigate one thing, get one report
- Smaller binary, fewer dependencies

**When to use which:**
- Use Tracee for system-wide security monitoring and detection
- Use procscope for focused investigation of a specific process or binary

### vs. Tetragon (Cilium/Isovalent)

**Tetragon** is a Kubernetes-first eBPF-based security observability and runtime enforcement tool.

**Where Tetragon is better:**
- Deep Kubernetes integration (CRDs, pod context)
- TracingPolicy for custom probe definitions
- Enforcement (kill, override) capabilities
- Process credential tracking
- Network observability at L3/L4/L7

**Where procscope is better:**
- Works without Kubernetes
- No YAML policies needed
- Immediate one-shot use
- Evidence bundle and report generation
- Simpler deployment on bare hosts

**When to use which:**
- Use Tetragon for K8s cluster security observability
- Use procscope for host-level process investigation

### vs. Inspektor Gadget

**Inspektor Gadget** is a collection of eBPF-based debugging tools for Kubernetes.

**Where Inspektor Gadget is better:**
- Wide variety of built-in gadgets
- Kubernetes-native (kubectl plugin)
- Network, DNS, and filesystem gadgets
- Container-aware by default

**Where procscope is better:**
- Works on any Linux host
- Process-scoped by design
- Evidence bundle output
- Single focused tool vs. gadget collection

### vs. strace

**strace** is the classic Linux syscall tracer using ptrace.

**Where strace is better:**
- Universally available (every Linux distro)
- No root needed for own processes
- Full syscall coverage (every syscall visible)
- Mature, well-understood output format
- Works on ancient kernels

**Where procscope is better:**
- Lower overhead (eBPF vs ptrace)
- No SIGSTOP (does not slow the target)
- Automatic fork following with tree tracking
- Structured output (JSON/JSONL)
- Evidence bundle with narrative summary
- Higher-level event classification (not raw syscall numbers)

**When to use which:**
- Use strace for quick syscall debugging, especially on minimal systems
- Use procscope for security investigations where you need structured, reportable output

### vs. sysdig

**sysdig** is a powerful system inspection tool using kernel modules or eBPF.

**Where sysdig is better:**
- Full system visibility
- Rich filtering (chisels, Lua scripting)
- Container-aware
- Both kernel module and eBPF backends
- Mature ecosystem

**Where procscope is better:**
- Process-scoped by default
- No kernel module option (eBPF only — safer)
- Evidence bundle output
- Simpler installation
- Focused on investigation use case

## Honest Assessment

procscope fills a specific niche: **immediate, process-scoped investigation on a Linux host without adopting a larger platform**. It is NOT a replacement for any of the tools above in their primary use cases.

If you need system-wide monitoring → use Tracee or sysdig.
If you need K8s security → use Tetragon.
If you need K8s debugging → use Inspektor Gadget.
If you need raw syscall visibility → use strace.
If you need to quickly investigate what a specific binary does → use procscope.
