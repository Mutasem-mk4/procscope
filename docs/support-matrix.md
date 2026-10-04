# Support Matrix

This document describes exactly what procscope can and cannot observe, under what conditions, and with what confidence.

## Kernel Requirements

| Requirement | Minimum | Recommended | Notes |
|------------|---------|-------------|-------|
| Kernel version | 5.8 | 6.1+ | Ring buffer, CAP_BPF |
| BTF | Required | Required | `CONFIG_DEBUG_INFO_BTF=y` |
| BPF | Required | Required | `CONFIG_BPF=y`, `CONFIG_BPF_SYSCALL=y` |
| Tracepoints | Required | Required | `CONFIG_TRACEPOINTS=y` |

### BTF Availability

BTF is available by default on:
- ✅ Kali Linux (6.x kernels)
- ✅ Parrot Security OS (6.x kernels)
- ✅ Ubuntu 20.10+
- ✅ Fedora 31+
- ✅ Debian 12+
- ✅ Arch Linux (rolling, current kernels)
- ⚠️ Older RHEL/CentOS may need explicit `kernel-debuginfo` or BTF backport

## Event Support Matrix

### Process Lifecycle

| Event | Syscall/Tracepoint | Confidence | Notes |
|-------|-------------------|------------|-------|
| exec | `sched/sched_process_exec` | **Exact** | Filename from tracepoint data |
| fork | `sched/sched_process_fork` | **Exact** | Child PID auto-tracked |
| exit | `sched/sched_process_exit` | **Exact** | Exit code best-effort |
| argv | N/A | **Partial** | Limited by eBPF stack; first args only |
| ppid | `task_struct->real_parent` | **Exact** | Via CO-RE |

### File Activity

| Event | Syscall | Confidence | Limitations |
|-------|---------|------------|-------------|
| open | `sys_enter_openat` | **Best-effort** | Misses `open()` (rare on modern libc), `openat2` |
| create | Inferred from openat flags | **Best-effort** | O_CREAT flag detection |
| rename | `sys_enter_renameat2` | **Best-effort** | Misses `rename()` (uses `renameat2` on modern libc) |
| unlink | `sys_enter_unlinkat` | **Best-effort** | Misses `unlink()` (uses `unlinkat` on modern libc) |
| chmod | `sys_enter_fchmodat` | **Best-effort** | Misses `chmod()`, `fchmod()` |
| chown | `sys_enter_fchownat` | **Best-effort** | Misses `chown()`, `fchown()`, `lchown()` |
| read/write | Not traced | **N/A** | Data content is not captured by design |
| io_uring | Not traced | **N/A** | io_uring file ops are invisible |
| sendfile | Not traced | **N/A** | — |

**Path resolution:** Paths from `openat` with `dirfd != AT_FDCWD` will be relative, not absolute. Full path resolution would require resolving `/proc/[pid]/fd/[dirfd]` which is expensive in BPF.

### Network Activity

| Event | Syscall | Confidence | Limitations |
|-------|---------|------------|-------------|
| connect | `sys_enter_connect` | **Best-effort** | IPv4/IPv6 address+port extracted |
| accept | `sys_enter_accept4` | **Best-effort** | Remote address not available at enter (would need exit probe) |
| bind | `sys_enter_bind` | **Best-effort** | Local address+port extracted |
| listen | `sys_enter_listen` | **Best-effort** | Backlog value captured |
| protocol | Inferred | **Inferred** | Always reports "tcp" (socket type not available at syscall enter) |
| UDP | Partially | **Best-effort** | `connect()` on UDP sockets visible, but `sendto()` is not traced |
| Unix domain | Partially | **Best-effort** | Address family visible but path not extracted |

### DNS

| Feature | Method | Confidence | Limitations |
|---------|--------|------------|-------------|
| DNS queries | Not directly | **N/A** | DNS extraction is NOT implemented in eBPF v1 |
| DNS via connect | `connect()` to port 53 | **Inferred** | Can detect connections to port 53, but not query content |

**Honest assessment:** Full DNS query extraction would require packet payload parsing in BPF (complex, fragile) or userspace socket buffer reading. This is deferred to post-v1 and will remain best-effort when implemented.

### Privilege Transitions

| Event | Syscall | Confidence | Limitations |
|-------|---------|------------|-------------|
| setuid | `sys_enter_setuid` | **Exact** | Old/new UID captured |
| setgid | `sys_enter_setgid` | **Exact** | Old/new GID captured |
| ptrace | `sys_enter_ptrace` | **Best-effort** | Request type + target PID |
| setresuid | Not traced | **N/A** | Would need additional probe |
| setresgid | Not traced | **N/A** | Would need additional probe |
| capabilities | Not traced | **N/A** | cap_set would need separate probe |

### Namespace Changes

| Event | Syscall | Confidence | Limitations |
|-------|---------|------------|-------------|
| setns | `sys_enter_setns` | **Best-effort** | NS type flags captured |
| unshare | `sys_enter_unshare` | **Best-effort** | Clone flags captured |
| clone (ns flags) | Not directly | **N/A** | Fork probe captures clone but doesn't parse ns flags |

### Mount Operations

| Event | Syscall | Confidence | Limitations |
|-------|---------|------------|-------------|
| mount | `sys_enter_mount` | **Best-effort** | Source, target, fstype, flags |
| umount | Not traced | **N/A** | — |
| mount_setattr | Not traced | **N/A** | — |

## What procscope Does NOT Observe

- **File read/write content** — by design, data payloads are not captured
- **io_uring operations** — completely invisible to syscall tracepoints
- **sendfile/splice** — not traced
- **Memory mapping** — mmap not traced
- **Signal delivery** — not traced (except as it causes exit)
- **IPC** — shared memory, message queues, semaphores not traced
- **eBPF program loading** — by other processes not traced
- **Kernel module loading** — not traced
- **Static binary internals** — statically linked programs may bypass libc wrappers; most syscalls are still visible via tracepoints
- **DoT/DoH DNS** — encrypted DNS is opaque
- **Container runtime metadata** — Docker/containerd labels require API access

## Architecture Support

| Architecture | Status | Notes |
|-------------|--------|-------|
| amd64 (x86_64) | ✅ Supported | Primary target |
| arm64 (aarch64) | ✅ Supported | Cross-compiled, less tested |
| arm (32-bit) | ❌ Not supported | — |
| riscv64 | ❌ Not supported | May work with BTF but untested |
