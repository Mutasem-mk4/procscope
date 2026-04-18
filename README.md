<p align="center">
  <img src="assets/header.jpeg" alt="procscope header banner" width="100%">
</p>

# procscope — eBPF Process Tracer for Linux Malware Triage & Incident Response

**Zero-overhead, zero-config eBPF process tracer for Linux.**
Trace malware behavior, investigate suspicious binaries, and audit container workloads — without `strace` overhead or the complexity of system-wide EDR daemons like Falco or Tetragon.

<p align="center">
  <a href="https://github.com/Mutasem-mk4/procscope/releases">
    <img src="https://img.shields.io/github/v/release/Mutasem-mk4/procscope?style=flat-square&label=release" alt="Latest Release">
  </a>
  <a href="https://github.com/Mutasem-mk4/procscope/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/Mutasem-mk4/procscope/ci.yml?branch=master&style=flat-square&label=ci" alt="CI Status">
  </a>
  <a href="https://github.com/Mutasem-mk4/procscope/actions/workflows/packaging-quality.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/Mutasem-mk4/procscope/packaging-quality.yml?branch=master&style=flat-square&label=packaging" alt="Packaging Status">
  </a>
  <a href="https://github.com/Mutasem-mk4/procscope/actions/workflows/security-suite.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/Mutasem-mk4/procscope/security-suite.yml?branch=master&style=flat-square&label=security" alt="Security Status">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/github/go-mod/go-version/Mutasem-mk4/procscope?style=flat-square&label=go" alt="Go Version">
  <img src="https://img.shields.io/github/license/Mutasem-mk4/procscope?style=flat-square&label=license" alt="License">
  <img src="https://img.shields.io/badge/linux-amd64%20%7C%20arm64-2F6FEB?style=flat-square" alt="Supported Linux Architectures">
  <img src="https://img.shields.io/badge/awesome--go-mentioned-1F883D?style=flat-square&logo=go" alt="Awesome Go Mentioned">
</p>

Launch a command under observation — or attach to an existing process — and see what it actually does at runtime: process lifecycle, file activity, network connections, privilege transitions, namespace changes, and more.

**Designed for:** security research, malware triage, reverse engineering support, incident response, and deep debugging.

**Not designed for:** EDR, SIEM, Kubernetes-first monitoring, policy enforcement, or whole-system tracing.

## Quick Start 

[![Try it in the Browser](https://img.shields.io/badge/Try_in_Browser-Killercoda-23C13F?style=for-the-badge&logoColor=white)](https://killercoda.com/mutasem04/scenario/procscope-scenario)

### 1-Minute Install (Go 1.25+)

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
| Arch / BlackArch package | `arch/PKGBUILD` maintained in-tree; not yet shipped by BlackArch |

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

**procscope** is a process-first local investigator. It is not an EDR, not a SIEM, and not a policy engine. It is designed to answer one question well: *what did this process actually do?*
