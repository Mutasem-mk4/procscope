<p align="center">
  <img src="assets/header.jpeg" alt="procscope header banner" width="100%">
</p>

# procscope — eBPF Process Tracer for Linux Malware Triage & Incident Response

**Zero-overhead, zero-config eBPF process tracer for Linux.**
Trace malware behavior, investigate suspicious binaries, and audit container workloads — without `strace` overhead or the complexity of system-wide EDR daemons like Falco or Tetragon.

<p align="center">
  <!-- Distribution & Recognition -->
  <a href="https://github.com/Mutasem-mk4/procscope/releases">
    <img src="https://img.shields.io/github/v/tag/Mutasem-mk4/procscope?style=flat-square&color=8A2BE2&label=release" alt="Latest Release">
  </a>
  <a href="https://blackarch.org/">
    <img src="https://img.shields.io/badge/BlackArch-000000?style=flat-square&logo=archlinux&logoColor=B00000" alt="BlackArch Linux">
  </a>
  <a href="https://github.com/avelino/awesome-go">
    <img src="https://img.shields.io/badge/Awesome--Go-Mentioned-15C213?style=flat-square&logo=go" alt="Awesome Go">
  </a>
  <a href="https://goreportcard.com/report/github.com/Mutasem-mk4/procscope">
    <img src="https://goreportcard.com/badge/github.com/Mutasem-mk4/procscope?style=flat-square" alt="Go Report Card">
  </a>
  <br>
  <!-- Quality & Security -->
  <img src="https://img.shields.io/github/actions/workflow/status/Mutasem-mk4/procscope/ci.yml?style=flat-square&label=CI" alt="CI Status">
  <img src="https://img.shields.io/github/actions/workflow/status/Mutasem-mk4/procscope/security-suite.yml?style=flat-square&label=security" alt="Security Suite Status">
  <img src="https://img.shields.io/github/license/Mutasem-mk4/procscope?style=flat-square&color=000000" alt="License">
  <img src="https://img.shields.io/github/stars/Mutasem-mk4/procscope?style=flat-square&color=F9A825" alt="GitHub Stars">
  <br>
  <!-- Engineering Core -->
  <img src="https://img.shields.io/badge/eBPF-Powered-blue?style=flat-square" alt="Powered by eBPF">
  <img src="https://img.shields.io/badge/Latency-%3C50%C2%B5s-blue?style=flat-square" alt="Latency">
  <img src="https://img.shields.io/badge/Heuristics-Enabled-orange?style=flat-square" alt="Heuristics Enabled">
</p>

Launch a command under observation — or attach to an existing process — and see what it actually does at runtime: process lifecycle, file activity, network connections, privilege transitions, namespace changes, and more.

**Designed for:** security research, malware triage, reverse engineering support, incident response, and deep debugging.

**Not designed for:** EDR, SIEM, Kubernetes-first monitoring, policy enforcement, or whole-system tracing.

## Quick Start 

[![Try it in the Browser](https://img.shields.io/badge/Try_in_Browser-Killercoda-23C13F?style=flat-square&logoColor=white)](https://killercoda.com/mutasem04/scenario/procscope-scenario)

### 1-Minute Install (Go 1.24+)

```bash
go install github.com/Mutasem-mk4/procscope/cmd/procscope@latest
procscope --version
```

```bash
# Trace a command
sudo procscope -- ./suspicious-binary

# Attach to a running process
sudo procscope -p 1234

# Save evidence bundle + Markdown report
sudo procscope --out case-001 --summary report.md -- ./installer.sh

# Stream events as JSONL
sudo procscope --jsonl events.jsonl -- ./tool
```

## What procscope Observes

| Category | Events | Confidence |
|----------|--------|------------|
| **Process lifecycle** | exec, fork/clone, exit (with codes) | Exact |
| **File activity** | open, rename, unlink, chmod, chown | Best-effort |
| **Network activity** | connect, accept, bind, listen (IP:port) | Best-effort |
| **Privilege transitions** | setuid, setgid, ptrace | Exact / Best-effort |
| **Namespace changes** | setns, unshare | Best-effort |
| **Mount operations** | mount | Best-effort |

> **Honesty note:** procscope does NOT claim to capture all process activity.
> See [docs/support-matrix.md](docs/support-matrix.md) for exact details on capabilities and blindspots.

## Requirements

- **Linux kernel 5.8+** with BTF (`CONFIG_DEBUG_INFO_BTF=y`)
- **Root** or `CAP_BPF` + `CAP_PERFMON` + `CAP_SYS_RESOURCE`
- **Architectures:** amd64, arm64

procscope will detect missing capabilities at startup and provide actionable guidance.

## Packaging Status

| Channel | Status |
|---------|--------|
| GitHub releases | Available |
| `go install` | Available |
| Debian / Kali / Parrot packages | Packaging metadata maintained in-tree; not yet shipped by the distro |
| Arch / BlackArch package | Available in BlackArch |

## Installation

Note: Running procscope usually requires `sudo` (eBPF capabilities).

### 1. Go Install

```bash
go install github.com/Mutasem-mk4/procscope/cmd/procscope@latest
```

### 2. Direct Download

Download the release asset that matches your architecture from:

- https://github.com/Mutasem-mk4/procscope/releases/latest

Current release assets include:

- Debian package (`.deb`)
- Linux tarballs for `amd64` and `arm64`

### 3. Build from Source

```bash
git clone https://github.com/Mutasem-mk4/procscope.git
cd procscope
make build
sudo install -m755 bin/procscope /usr/local/bin/procscope
```

### 4. Native Package Managers

These commands are the target install experience after distro acceptance.

**BlackArch Linux:**
```bash
sudo pacman -S procscope
```

**Kali Linux & Parrot OS:**
```bash
sudo apt update && sudo apt install procscope
```

## Output Formats

### Live Timeline

Compact, color-coded terminal output during investigation:

```
TIME         PID   COMM            EVENT              DETAILS
[+    0ms]   1234  suspicious      process.exec       /tmp/suspicious-binary
[+   12ms]   1234  suspicious      file.open          /etc/passwd [read]
[+   15ms]   1234  suspicious      net.connect        ipv4 → 93.184.216.34:443
[+   18ms] ! 1234  suspicious      priv.setuid        uid 1000 → 0
[+   20ms]   1235  sh              process.exec       /bin/sh
[+   25ms]   1235  sh              process.exit        exit_code=0
[+   30ms]   1234  suspicious      process.exit        exit_code=0
```

### JSONL Event Stream

Machine-readable, one event per line:

```bash
procscope --jsonl events.jsonl -- ./command
```

### Evidence Bundle

Structured directory for incident response:

```
case-001/
├── metadata.json       # Investigation metadata
├── events.jsonl        # Complete event stream
├── process-tree.txt    # Human-readable process tree
├── files.json          # File activity summary
├── network.json        # Network activity summary
├── notable.json        # Security-relevant events
└── summary.md          # Markdown executive summary
```

### Markdown Summary

Team-ready report with overview, process tree, event breakdown, file/network activity tables, notable events, and honest limitations.

## Configuration & Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--pid` | `-p` | Attach to existing PID | — |
| `--name` | `-n` | Attach by process name | — |
| `--out` | `-o` | Evidence bundle directory | — |
| `--jsonl` | | JSONL output file | — |
| `--summary` | | Markdown summary file | — |
| `--no-color` | | Disable ANSI colors | false |
| `--quiet` | `-q` | Suppress live timeline | false |
| `--max-args` | | Max argv elements | 64 |
| `--max-path` | | Max path string length | 4096 |
| `--skip-checks` | | Skip privilege checks | false |

## Safe Defaults

- **No environment dumping** — env vars are not captured by default
- **No secret capture** — payload/body content is not traced
- **Bounded lengths** — arguments and paths are truncated at configurable limits
- **Pattern-based redaction** — values matching `password`, `token`, `secret`, etc. are redacted

## Architecture

```
┌───────────────────────────────────────┐
│              CLI (cobra)              │
├──────────┬────────────┬───────────────┤
│ Launcher │  Attacher  │  Cap Check    │
├──────────┴────────────┴───────────────┤
│           Event Correlator            │
│   (process tree, investigation ID)    │
├───────────────────────────────────────┤
│          eBPF Tracer Manager          │
│   (load, attach, ring buffer read)    │
├───────────────────────────────────────┤
│        eBPF Programs (kernel)         │
│  tracepoints: sched, syscalls, etc.   │
├───────────────────────────────────────┤
│            Output Layer               │
│  timeline │ JSON │ bundle │ summary   │
└───────────────────────────────────────┘
```

See [docs/architecture.md](docs/architecture.md) for detailed design.

## Comparison with Other Tools

| Feature | procscope | Tracee | Tetragon | Inspektor Gadget | strace |
|---------|-----------|--------|----------|------------------|--------|
| **Focus** | Process-scoped investigation | Runtime security | K8s observability | K8s debugging | Syscall tracing |
| **Scope** | Single process tree | System-wide | System/pod-wide | System/pod-wide | Single process |
| **Setup** | Zero config | Policy config | CRDs | kubectl | Zero config |
| **Evidence bundle** | ✓ | ✗ | ✗ | ✗ | ✗ |
| **Markdown report** | ✓ | ✗ | ✗ | ✗ | ✗ |
| **Process tree** | ✓ auto-follows forks | ✓ | ✓ | ✓ | `-f` flag |
| **K8s-native** | ✗ | ✓ | ✓ | ✓ | ✗ |
| **Policy engine** | ✗ | ✓ | ✓ | ✗ | ✗ |

See [docs/comparison.md](docs/comparison.md) for honest, detailed comparison.

## Documentation

- [Building from Source](BUILDING.md)
- [Architecture](docs/architecture.md)
- [Support Matrix](docs/support-matrix.md)
- [Acceptance Risk Assessment](docs/acceptance-risk.md)
- [Security Model](docs/security-model.md)
- [Privacy Model](docs/privacy-model.md)
- [Packaging Guide](docs/packaging.md)
- [Distribution Submission Playbook](docs/packaging-submission-playbook.md)
- [Comparison](docs/comparison.md)
- [Design Decisions](docs/design-decisions/)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).
`procscope` is heavily community-driven; check issues labeled `good-first-issue` to get started quickly.

## Security

See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## Community

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community participation guidelines.

## License

[MIT](LICENSE)

---

## About the Author

**procscope** was developed by [Mutasem Kharma (معتصم خرما)](https://github.com/Mutasem-mk4), a **Security Engineer** and **eBPF Specialist** focused on building high-performance, offensive and defensive security tools for the modern cloud-native landscape. Mutasem specializes in low-level system observation, automated vulnerability research, and autonomous security agents.

---

**procscope** is a process-first local investigator. It is not an EDR, not a SIEM, and not a policy engine. It is designed to answer one question well: *what did this process actually do-*
