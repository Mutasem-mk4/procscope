# Privacy Model

## Data Collection Scope

procscope collects process behavior metadata. It does NOT collect:
- File content
- Network payload data
- User input/keystrokes
- Screen content
- Environment variables (by default)

## What Can Appear in Output

| Data Type | Present in Output? | Redaction |
|-----------|-------------------|-----------|
| Process ID (PID) | Yes | Not redacted |
| Process name (comm) | Yes | Not redacted |
| Command-line arguments | Yes (bounded) | Sensitive patterns auto-redacted |
| File paths | Yes (bounded) | Sensitive patterns auto-redacted |
| IP addresses | Yes | Not redacted |
| Port numbers | Yes | Not redacted |
| UID/GID values | Yes | Not redacted |
| Environment variables | No (by default) | Opt-in only |
| File contents | No | Never captured |
| Network payloads | No | Never captured |
| DNS query names | Planned (best-effort) | Not redacted |

## Redaction Controls

procscope automatically redacts values matching these patterns (case-insensitive):
- `password`, `passwd`
- `secret`
- `token`
- `api_key`, `apikey`, `api-key`
- `authorization`
- `credential`
- `private_key`, `private-key`

Example: `--api_key=abc123` → `[REDACTED]`

## Evidence Bundle Privacy

Evidence bundles (created with `--out`) contain machine-readable investigation data. Before sharing:

1. **Review** the bundle contents, especially `events.jsonl`
2. **Redact** any sensitive information not caught by auto-redaction
3. **Restrict** file permissions — bundles are created with 0750/0640
4. **Consider** that IP addresses, file paths, and command names may be sensitive in your context

## Data Retention

procscope does NOT:
- Persist any data beyond the current investigation
- Write to system logs
- Send data to external services
- Cache data between runs

All output is explicitly requested by the operator via `--out`, `--jsonl`, or `--summary` flags.

## Multi-User Considerations

- procscope running as root can observe any process on the system
- When using `CAP_BPF` without root, it can only observe own-user processes (unless `CAP_SYS_PTRACE` is granted)
- Evidence bundles are created with the running user's ownership
- In shared systems, operators should be aware that traced command lines may be visible in the bundle
