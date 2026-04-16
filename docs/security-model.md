# Security Model

## Privilege Architecture

procscope requires elevated privileges to load eBPF programs and attach kernel tracepoints. This is fundamental to how eBPF works on Linux.

### Minimum Capabilities

| Capability | Why | Alternative |
|-----------|-----|-------------|
| `CAP_BPF` | Load BPF programs into kernel | `CAP_SYS_ADMIN` (broader, legacy) |
| `CAP_PERFMON` | Attach to tracepoints and perf events | `CAP_SYS_ADMIN` (broader, legacy) |
| `CAP_SYS_RESOURCE` | Raise RLIMIT_MEMLOCK for BPF maps | `ulimit -l unlimited` before running |

### Optional Capabilities

| Capability | Why |
|-----------|-----|
| `CAP_SYS_PTRACE` | To attach to processes owned by other users |

### Running Without Root

```bash
# Grant minimum capabilities to the binary
sudo setcap cap_bpf,cap_perfmon,cap_sys_resource+ep /usr/bin/procscope

# Then run without sudo (own processes only)
procscope -- ./my-program
```

Note: `CAP_SYS_PTRACE` is additionally needed for `-p <other-user's-PID>`.

## What procscope Does NOT Do

- ❌ **Does not modify system security policy**
- ❌ **Does not set capabilities in package install scripts** — this is left to the administrator
- ❌ **Does not modify kernel parameters**
- ❌ **Does not persist kernel state** — all BPF programs are cleaned up on exit
- ❌ **Does not run as a daemon** — it is a user-invoked tool
- ❌ **Does not send data externally** — no telemetry, no analytics, no network calls
- ❌ **Does not enforce policy** — it only observes and reports
- ❌ **Does not intercept or block syscalls** — observation only

## BPF Program Safety

All procscope BPF programs:
- Are verified by the kernel's BPF verifier before loading
- Cannot crash the kernel (verifier guarantee)
- Cannot access arbitrary kernel memory (verifier guarantee)
- Use bounded loops and bounded map access
- Are automatically cleaned up when procscope exits
- Are not pinned to the BPF filesystem (no persistence)

## Data Handling

### What Is Captured

- Process metadata: PID, PPID, comm, filename, arguments (bounded)
- File paths (bounded, no content)
- Network addresses and ports (no payload)
- Privilege transition metadata (UID/GID values)
- Namespace operation flags
- Mount metadata (source, target, fstype)

### What Is NOT Captured

- Environment variables (by default)
- File content / read data / write data
- Network payload / packet data
- Memory contents
- Encryption keys or secrets (unless in argv — see redaction)

### Safe Defaults

1. **No environment dumping** — `ShowEnv` is false by default
2. **Bounded arguments** — max 64 args, max 1024 chars each
3. **Bounded paths** — max 4096 chars
4. **Sensitive pattern redaction** — values matching patterns like `password`, `token`, `secret`, `api_key` are replaced with `[REDACTED]`
5. **Restricted output permissions** — evidence bundles use 0750/0640

### Ethical Considerations

procscope is designed for **authorized security research** on systems where the operator has legitimate access. It should be used:

- ✅ On systems you own or have authorization to test
- ✅ For malware analysis in controlled environments
- ✅ For incident response on compromised systems you administer
- ✅ For debugging your own applications
- ❌ Not for unauthorized surveillance
- ❌ Not for monitoring users without consent
- ❌ Not for circumventing security controls

The tool itself is neutral — it provides visibility that root already has. The ethical boundary is in how it is used.
