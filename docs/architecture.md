# Architecture

## Overview

procscope is a process-scoped runtime investigation tool built in Go with eBPF for kernel-level observation. The architecture follows a pipeline model:

```
Target Process → eBPF Probes → Ring Buffer → Event Parser → Correlator → Output Sinks
```

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         User Space                                  │
│                                                                     │
│  ┌──────────┐    ┌─────────────┐    ┌──────────────┐               │
│  │   CLI    │───▶│  Launcher/  │───▶│   Correlator │               │
│  │ (cobra)  │    │  Attacher   │    │ (process tree│               │
│  └────┬─────┘    └─────────────┘    │  event enrich│               │
│       │                              └──────┬───────┘               │
│       │                                     │                       │
│       ▼                                     ▼                       │
│  ┌──────────┐                        ┌──────────────┐              │
│  │   Caps   │                        │   Output     │              │
│  │  Check   │                        │   Sinks      │              │
│  └──────────┘                        │ ┌──────────┐ │              │
│                                      │ │ Timeline │ │              │
│  ┌──────────────────┐                │ ├──────────┤ │              │
│  │  Tracer Manager  │────events─────▶│ │  JSONL   │ │              │
│  │  (load, attach,  │                │ ├──────────┤ │              │
│  │   read ringbuf)  │                │ │  Bundle  │ │              │
│  └────────┬─────────┘                │ ├──────────┤ │              │
│           │                          │ │ Summary  │ │              │
│           │                          │ └──────────┘ │              │
├───────────┼──────────────────────────┴──────────────┴──────────────┤
│           │              Kernel Space                               │
│           ▼                                                         │
│  ┌──────────────────────────────────────────────────────┐          │
│  │                eBPF Programs                          │          │
│  │                                                       │          │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │          │
│  │  │ Process      │  │ File         │  │ Network    │ │          │
│  │  │ sched/exec   │  │ sys/openat   │  │ sys/connect│ │          │
│  │  │ sched/fork   │  │ sys/rename   │  │ sys/accept │ │          │
│  │  │ sched/exit   │  │ sys/unlinkat │  │ sys/bind   │ │          │
│  │  │              │  │ sys/fchmodat │  │ sys/listen │ │          │
│  │  └──────┬───────┘  └──────┬───────┘  └─────┬──────┘ │          │
│  │         │                 │                 │        │          │
│  │  ┌──────────────┐  ┌──────────────┐                  │          │
│  │  │ Privilege    │  │ Namespace    │                  │          │
│  │  │ sys/setuid   │  │ sys/setns    │                  │          │
│  │  │ sys/setgid   │  │ sys/unshare  │                  │          │
│  │  │ sys/ptrace   │  │ sys/mount    │                  │          │
│  │  └──────┬───────┘  └──────┬───────┘                  │          │
│  │         │                 │                           │          │
│  │         ▼                 ▼                           │          │
│  │  ┌────────────────────────────────┐                  │          │
│  │  │       BPF Ring Buffer          │                  │          │
│  │  │   (512KB, shared events map)   │                  │          │
│  │  └────────────────────────────────┘                  │          │
│  │                                                       │          │
│  │  ┌────────────────────────────────┐                  │          │
│  │  │     tracked_pids Hash Map      │                  │          │
│  │  │   (PID filtering per-event)    │                  │          │
│  │  └────────────────────────────────┘                  │          │
│  └──────────────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
cmd/procscope/       → Entry point
internal/
├── cli/             → Cobra command setup, flag parsing, run orchestration
├── tracer/          → eBPF program loading, attachment, ring buffer reading
├── events/          → Event types, schema, and correlator (process tree)
├── process/         → Process launcher, PID attacher, /proc tree builder
├── output/          → Timeline, JSONL, evidence bundle, Markdown summary
├── caps/            → Runtime capability/privilege detection
├── redact/          → Safe-default redaction controls
└── version/         → Build-time version embedding
bpf/
├── procscope.c      → eBPF C program with all probes
└── headers/
    └── vmlinux.h    → Minimal kernel type subset for CO-RE
```

## Data Flow

1. **Initialization:** CLI parses flags, checks privileges, loads eBPF programs
2. **Target setup:** Launch new process or discover existing PID's children
3. **PID registration:** Target PIDs added to `tracked_pids` BPF hash map
4. **Event capture:** eBPF probes fire on syscalls/tracepoints, check `tracked_pids`, submit to ring buffer
5. **Event reading:** Go goroutine reads from ring buffer, parses binary events to Go structs
6. **Correlation:** Events enriched with investigation ID, tracked in process tree, fork children auto-tracked
7. **Output:** Events dispatched to enabled output sinks (timeline, JSONL, collector for bundle)
8. **Completion:** On process exit or Ctrl+C, evidence bundle and summary are written

## Key Design Decisions

| Decision | Rationale | ADR |
|----------|-----------|-----|
| eBPF over ptrace | Lower overhead, no SIGSTOP, kernel-level visibility | [001](design-decisions/001-ebpf-over-ptrace.md) |
| cilium/ebpf library | Go-native, no CGO, well-maintained, CO-RE support | [002](design-decisions/002-cilium-ebpf.md) |
| No CGO | Simpler cross-compilation, static binaries, fewer build deps | [003](design-decisions/003-no-cgo.md) |
| Versioned event schema | Forward-compatible consumers, stable machine-readable output | [004](design-decisions/004-event-schema.md) |
| Ring buffer over perf buffer | Lower overhead, better ordering, requires kernel 5.8+ | Kernel requirement is acceptable for target distros |
| PID-scoped filtering in BPF | Reduces userspace event volume, focuses on investigation target | Core design principle |
| Channel-based event pipeline | Natural Go concurrency, bounded backpressure | Standard Go pattern |

## Concurrency Model

- **Ring buffer reader:** Single goroutine reading from BPF ring buffer
- **Event consumer:** Single goroutine consuming from correlator channel
- **Correlator:** Thread-safe with RWMutex for concurrent access
- **Output sinks:** Write-locked where needed (JSONL writer)
- **Signal handling:** Context cancellation propagates to all goroutines

## Error Handling

- eBPF load failures: Fatal with actionable diagnostics
- Individual tracepoint failures: Non-fatal, logged as warning (graceful degradation)
- Ring buffer overflow: Events silently dropped (bounded-loss design)
- Correlator channel full: Events dropped (backpressure safety)
- Output write errors: Logged, non-fatal during investigation
