# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in procscope, please report it responsibly:

1. **Do NOT open a public GitHub issue.**
2. Email: **security@procscope.dev** (or use GitHub's private vulnerability reporting if available).
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Impact assessment
   - Suggested fix (if any)

We will acknowledge receipt within 48 hours and aim to provide an initial assessment within 5 business days.

## Scope

The following are in scope for security reports:

- Privilege escalation via procscope
- Unintended data exposure (secrets, env vars, etc.)
- eBPF program escape or misuse
- Denial of service via resource exhaustion
- Supply chain vulnerabilities in dependencies

## Threat Model

procscope is a **privileged tool** designed to observe process behavior. By design, it requires elevated privileges (root or specific capabilities). The threat model acknowledges:

### In Scope

- **Accidental data leakage:** procscope could inadvertently capture and expose sensitive data (passwords, tokens, keys) in its output. Mitigated by safe defaults and redaction controls.
- **Output file permissions:** Evidence bundles must not be world-readable. Mitigated by restrictive file permissions (0640/0750).
- **Dependency supply chain:** Third-party Go modules could introduce vulnerabilities. Mitigated by minimal dependency count and `govulncheck`.

### Accepted Risks (by design)

- **Root/CAP_BPF access:** procscope requires kernel-level access. A user with root access already has full system control. procscope does not add new attack surface beyond what root already provides.
- **Process observation:** By design, procscope observes process behavior. This is its stated purpose for authorized use.

### Out of Scope

- Kernel vulnerabilities in the eBPF subsystem itself
- Social engineering attacks
- Physical access attacks

## Privilege Requirements

| Capability | Purpose | Required? |
|-----------|---------|-----------|
| `CAP_BPF` | Load eBPF programs | Yes |
| `CAP_PERFMON` | Attach tracepoint probes | Yes |
| `CAP_SYS_RESOURCE` | eBPF map memory allocation | Yes |
| `CAP_SYS_ADMIN` | Legacy fallback (pre-5.8 kernels) | Alternative |
| `CAP_SYS_PTRACE` | Attach to other users' processes | Optional |

procscope **never**:
- Sets capabilities in package scripts
- Modifies system security configuration
- Runs as a persistent daemon (unless explicitly opted in)
- Sends data to external services

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.1.x   | ✅ Current |

## Hardening Recommendations

1. Run with minimum necessary capabilities instead of full root:
   ```bash
   sudo setcap cap_bpf,cap_perfmon,cap_sys_resource+ep /usr/bin/procscope
   ```
2. Restrict evidence bundle directory permissions
3. Use `--max-args` and `--max-path` to limit data capture
4. Review evidence bundles before sharing
