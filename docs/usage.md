# Usage and Output Formats

This document describes how to configure `procscope` and the various output formats it provides.

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
