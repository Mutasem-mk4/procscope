// SPDX-License-Identifier: MIT OR GPL-2.0-only
//go:build ignore

/*
 * procscope.c — eBPF probes for process-scoped runtime investigation.
 *
 * This program hooks into kernel tracepoints to observe:
 *   - Process lifecycle (exec, fork, exit)
 *   - File activity (openat, rename, unlink, chmod, chown)
 *   - Network activity (connect, accept, bind, listen)
 *   - Privilege transitions (setuid, setgid, ptrace)
 *   - Namespace changes (setns, unshare)
 *   - Mount operations
 *
 * Design decisions:
 *   - Uses BPF ring buffer (requires kernel 5.8+)
 *   - Tracks PIDs via hash map, auto-tracking forked children
 *   - CO-RE for cross-kernel portability
 *   - All path/arg captures are bounded to prevent stack overflow
 *
 * Limitations:
 *   - Cannot capture full argv for exec (stack size limits)
 *   - File paths from openat use dirfd-relative paths when FD != AT_FDCWD
 *   - DNS extraction is NOT implemented in eBPF (done in userspace, best-effort)
 *   - Static binaries may not trigger expected syscall probes
 */

#include "headers/vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_core_read.h>

/*
 * The bundled vmlinux.h is intentionally minimal, so define the small subset
 * of BPF map and update constants this program needs when kernel headers don't.
 */
#ifndef BPF_MAP_TYPE_HASH
#define BPF_MAP_TYPE_HASH 1
#endif

#ifndef BPF_MAP_TYPE_RINGBUF
#define BPF_MAP_TYPE_RINGBUF 27
#endif

#ifndef BPF_ANY
#define BPF_ANY 0
#endif

char __license[] SEC("license") = "Dual MIT/GPL";

/* --- Constants --- */

#define MAX_PATH_LEN   256
#define MAX_ARGS_LEN   256
#define MAX_TRACKED    8192
#define TASK_COMM_LEN  16

/* AF_INET / AF_INET6 */
#define AF_INET  2
#define AF_INET6 10

/* --- Event type enum (must match Go side) --- */

enum event_type {
	EVENT_EXEC        = 1,
	EVENT_FORK        = 2,
	EVENT_EXIT        = 3,
	EVENT_FILE_OPEN   = 10,
	EVENT_FILE_RENAME = 12,
	EVENT_FILE_UNLINK = 13,
	EVENT_FILE_CHMOD  = 14,
	EVENT_FILE_CHOWN  = 15,
	EVENT_NET_CONNECT = 20,
	EVENT_NET_ACCEPT  = 21,
	EVENT_NET_BIND    = 22,
	EVENT_NET_LISTEN  = 23,
	EVENT_PRIV_SETUID = 30,
	EVENT_PRIV_SETGID = 31,
	EVENT_PRIV_PTRACE = 32,
	EVENT_NS_SETNS    = 40,
	EVENT_NS_UNSHARE  = 41,
	EVENT_MOUNT       = 50,
};

/* --- Event structure sent to userspace --- */

struct event {
	/* Common header */
	__u64 timestamp;
	__u32 event_type;
	__u32 pid;
	__u32 tid;
	__u32 ppid;
	__u64 cgroup_id;
	__u32 uid;
	__u32 gid;
	char  comm[TASK_COMM_LEN];

	/* Process fields */
	__u32 exit_code;
	__u32 child_pid;
	char  filename[MAX_PATH_LEN];

	/* File fields */
	__u32 flags;
	__u32 mode;
	char  path[MAX_PATH_LEN];
	char  path2[MAX_PATH_LEN];  /* newpath for rename, target for mount */

	/* Network fields */
	__u32 af;           /* address family */
	__u16 sport;
	__u16 dport;
	__u8  saddr[16];   /* IPv4 in first 4 bytes, or full IPv6 */
	__u8  daddr[16];
	__u32 protocol;
	__u32 backlog;      /* for listen */

	/* Privilege fields */
	__u32 old_uid;
	__u32 new_uid;
	__u32 old_gid;
	__u32 new_gid;
	__u64 ptrace_request;
	__u32 target_pid;

	/* Namespace fields */
	__u32 ns_type;
	__u64 clone_flags;

	/* Mount fields */
	char  fstype[64];
	__u64 mount_flags;

	/* Return value */
	__s32 retval;
	__u32 _pad;
};

/* --- Maps --- */

/* Ring buffer for events to userspace */
struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 512 * 1024); /* 512KB ring buffer */
} events SEC(".maps");

/* Hash map of tracked PIDs: key=pid, value=1 */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_TRACKED);
	__type(key, __u32);
	__type(value, __u8);
} tracked_pids SEC(".maps");

/* --- Helpers --- */

/* Check if the current process is being tracked */
static __always_inline int is_tracked(void) {
	__u32 pid = bpf_get_current_pid_tgid() >> 32;
	return bpf_map_lookup_elem(&tracked_pids, &pid) != NULL;
}

/* Fill common event header fields */
static __always_inline void fill_header(struct event *e, __u32 event_type) {
	__u64 pid_tgid = bpf_get_current_pid_tgid();
	__u64 uid_gid  = bpf_get_current_uid_gid();

	e->timestamp  = bpf_ktime_get_ns();
	e->event_type = event_type;
	e->pid        = pid_tgid >> 32;       /* tgid = userspace PID */
	e->tid        = (__u32)pid_tgid;       /* pid  = userspace TID */
	e->uid        = (__u32)uid_gid;
	e->gid        = uid_gid >> 32;
	e->cgroup_id  = bpf_get_current_cgroup_id();

	bpf_get_current_comm(&e->comm, sizeof(e->comm));

	/* Best-effort PPID read via current task */
	struct task_struct *task = (struct task_struct *)bpf_get_current_task();
	if (task) {
		struct task_struct *parent = NULL;
		bpf_probe_read_kernel(&parent, sizeof(parent), &task->real_parent);
		if (parent) {
			bpf_probe_read_kernel(&e->ppid, sizeof(e->ppid), &parent->tgid);
		}
	}
}

/* Track a child PID (add to tracked_pids map) */
static __always_inline void track_pid(__u32 pid) {
	__u8 val = 1;
	bpf_map_update_elem(&tracked_pids, &pid, &val, BPF_ANY);
}

/* ======================================================================
 * PROCESS LIFECYCLE PROBES
 * ====================================================================== */

SEC("tracepoint/sched/sched_process_exec")
int handle_exec(struct trace_event_raw_sched_process_exec *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_EXEC);

	/* Read filename from tracepoint data_loc encoding */
	unsigned int data_loc = ctx->__data_loc_filename;
	unsigned short off = data_loc & 0xFFFF;
	bpf_probe_read_kernel_str(e->filename, sizeof(e->filename),
		(void *)ctx + off);

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/sched/sched_process_fork")
int handle_fork(struct trace_event_raw_sched_process_fork *ctx) {
	/* Check if parent is tracked */
	__u32 parent_pid = ctx->parent_pid;
	if (bpf_map_lookup_elem(&tracked_pids, &parent_pid) == NULL)
		return 0;

	/* Auto-track the child */
	__u32 child_pid = ctx->child_pid;
	track_pid(child_pid);

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_FORK);
	e->child_pid = child_pid;

	/* Override PID to be the parent for this event */
	e->pid = parent_pid;

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/sched/sched_process_exit")
int handle_exit(struct trace_event_raw_sched_process_template *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_EXIT);

	/* Read exit code from task_struct */
	struct task_struct *task = (struct task_struct *)bpf_get_current_task();
	if (task) {
		int exit_code = 0;
		bpf_probe_read_kernel(&exit_code, sizeof(exit_code),
			(void *)task + __builtin_offsetof(struct task_struct, pid) - sizeof(int));
		/* Note: exit code extraction is best-effort; the exact offset
		 * depends on kernel version. CO-RE handles this via BTF. */
	}

	bpf_ringbuf_submit(e, 0);
	return 0;
}

/* ======================================================================
 * FILE ACTIVITY PROBES
 * ====================================================================== */

SEC("tracepoint/syscalls/sys_enter_openat")
int handle_openat(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_FILE_OPEN);

	/* args: dfd, filename, flags, mode */
	const char *filename = (const char *)ctx->args[1];
	__u32 flags = (__u32)ctx->args[2];
	__u32 mode  = (__u32)ctx->args[3];

	bpf_probe_read_user_str(e->path, sizeof(e->path), filename);
	e->flags = flags;
	e->mode  = mode;

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_renameat2")
int handle_rename(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_FILE_RENAME);

	/* args: olddfd, oldname, newdfd, newname, flags */
	const char *oldname = (const char *)ctx->args[1];
	const char *newname = (const char *)ctx->args[3];

	bpf_probe_read_user_str(e->path, sizeof(e->path), oldname);
	bpf_probe_read_user_str(e->path2, sizeof(e->path2), newname);

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_unlinkat")
int handle_unlink(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_FILE_UNLINK);

	/* args: dfd, pathname, flag */
	const char *pathname = (const char *)ctx->args[1];
	bpf_probe_read_user_str(e->path, sizeof(e->path), pathname);

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_fchmodat")
int handle_chmod(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_FILE_CHMOD);

	/* args: dfd, filename, mode */
	const char *filename = (const char *)ctx->args[1];
	__u32 mode = (__u32)ctx->args[2];

	bpf_probe_read_user_str(e->path, sizeof(e->path), filename);
	e->mode = mode;

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_fchownat")
int handle_chown(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_FILE_CHOWN);

	/* args: dfd, filename, user, group, flag */
	const char *filename = (const char *)ctx->args[1];
	__u32 user  = (__u32)ctx->args[2];
	__u32 group = (__u32)ctx->args[3];

	bpf_probe_read_user_str(e->path, sizeof(e->path), filename);
	e->new_uid = user;
	e->new_gid = group;

	bpf_ringbuf_submit(e, 0);
	return 0;
}

/* ======================================================================
 * NETWORK ACTIVITY PROBES
 * ====================================================================== */

SEC("tracepoint/syscalls/sys_enter_connect")
int handle_connect(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_NET_CONNECT);

	/* args: fd, uservaddr, addrlen */
	struct sockaddr *addr = (struct sockaddr *)ctx->args[1];

	__u16 family = 0;
	bpf_probe_read_user(&family, sizeof(family), &addr->sa_family);
	e->af = family;

	if (family == AF_INET) {
		struct sockaddr_in sin = {};
		bpf_probe_read_user(&sin, sizeof(sin), addr);
		e->dport = __builtin_bswap16(sin.sin_port);
		__builtin_memcpy(e->daddr, &sin.sin_addr.s_addr, 4);
	} else if (family == AF_INET6) {
		struct sockaddr_in6 sin6 = {};
		bpf_probe_read_user(&sin6, sizeof(sin6), addr);
		e->dport = __builtin_bswap16(sin6.sin6_port);
		__builtin_memcpy(e->daddr, &sin6.sin6_addr, 16);
	}

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_accept4")
int handle_accept(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_NET_ACCEPT);

	/* Accept details are populated on sys_exit, but we record the
	 * attempt here. Full address extraction on accept exit would
	 * require a separate sys_exit probe with per-CPU temp storage. */

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_bind")
int handle_bind(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_NET_BIND);

	/* args: fd, umyaddr, addrlen */
	struct sockaddr *addr = (struct sockaddr *)ctx->args[1];

	__u16 family = 0;
	bpf_probe_read_user(&family, sizeof(family), &addr->sa_family);
	e->af = family;

	if (family == AF_INET) {
		struct sockaddr_in sin = {};
		bpf_probe_read_user(&sin, sizeof(sin), addr);
		e->sport = __builtin_bswap16(sin.sin_port);
		__builtin_memcpy(e->saddr, &sin.sin_addr.s_addr, 4);
	} else if (family == AF_INET6) {
		struct sockaddr_in6 sin6 = {};
		bpf_probe_read_user(&sin6, sizeof(sin6), addr);
		e->sport = __builtin_bswap16(sin6.sin6_port);
		__builtin_memcpy(e->saddr, &sin6.sin6_addr, 16);
	}

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_listen")
int handle_listen(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_NET_LISTEN);

	/* args: fd, backlog */
	e->backlog = (__u32)ctx->args[1];

	bpf_ringbuf_submit(e, 0);
	return 0;
}

/* ======================================================================
 * PRIVILEGE PROBES
 * ====================================================================== */

SEC("tracepoint/syscalls/sys_enter_setuid")
int handle_setuid(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_PRIV_SETUID);

	e->old_uid = e->uid; /* current UID from header */
	e->new_uid = (__u32)ctx->args[0];

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_setgid")
int handle_setgid(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_PRIV_SETGID);

	e->old_gid = e->gid;
	e->new_gid = (__u32)ctx->args[0];

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_ptrace")
int handle_ptrace(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_PRIV_PTRACE);

	/* args: request, pid, addr, data */
	e->ptrace_request = ctx->args[0];
	e->target_pid = (__u32)ctx->args[1];

	bpf_ringbuf_submit(e, 0);
	return 0;
}

/* ======================================================================
 * NAMESPACE PROBES
 * ====================================================================== */

SEC("tracepoint/syscalls/sys_enter_setns")
int handle_setns(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_NS_SETNS);

	/* args: fd, nstype */
	e->ns_type = (__u32)ctx->args[1];

	bpf_ringbuf_submit(e, 0);
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_unshare")
int handle_unshare(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_NS_UNSHARE);

	e->clone_flags = ctx->args[0];

	bpf_ringbuf_submit(e, 0);
	return 0;
}

/* ======================================================================
 * MOUNT PROBE
 * ====================================================================== */

SEC("tracepoint/syscalls/sys_enter_mount")
int handle_mount(struct trace_event_raw_sys_enter *ctx) {
	if (!is_tracked())
		return 0;

	struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
	if (!e)
		return 0;

	__builtin_memset(e, 0, sizeof(*e));
	fill_header(e, EVENT_MOUNT);

	/* args: dev_name, dir_name, type, flags, data */
	const char *source = (const char *)ctx->args[0];
	const char *target = (const char *)ctx->args[1];
	const char *fstype = (const char *)ctx->args[2];
	__u64 mflags = ctx->args[3];

	bpf_probe_read_user_str(e->path, sizeof(e->path), source);
	bpf_probe_read_user_str(e->path2, sizeof(e->path2), target);
	bpf_probe_read_user_str(e->fstype, sizeof(e->fstype), fstype);
	e->mount_flags = mflags;

	bpf_ringbuf_submit(e, 0);
	return 0;
}
