/* SPDX-License-Identifier: MIT OR GPL-2.0-only */
/*
 * vmlinux.h — Minimal kernel type subset for procscope eBPF programs.
 *
 * This file provides the minimum kernel type definitions required for
 * procscope's eBPF probes to compile with CO-RE (Compile Once Run Everywhere).
 *
 * For full BTF-based vmlinux.h, generate from your running kernel:
 *   bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h
 *
 * This subset avoids bundling a multi-megabyte vmlinux.h and keeps the
 * build reproducible across kernel versions via CO-RE relocations.
 */

#ifndef __VMLINUX_H__
#define __VMLINUX_H__

/* Prevent system headers from conflicting when included with bpf helpers */
#pragma clang attribute push (__attribute__((preserve_access_index)), apply_to = record)

/* Basic types */
typedef unsigned char __u8;
typedef short int __s16;
typedef short unsigned int __u16;
typedef int __s32;
typedef unsigned int __u32;
typedef long long int __s64;
typedef long long unsigned int __u64;
typedef __u8 u8;
typedef __s16 s16;
typedef __u16 u16;
typedef __s32 s32;
typedef __u32 u32;
typedef __s64 s64;
typedef __u64 u64;
typedef __u16 __le16;
typedef __u16 __be16;
typedef __u32 __be32;
typedef __u64 __be64;
typedef __u32 __wsum;
typedef int pid_t;
typedef unsigned int uid_t;
typedef unsigned int gid_t;
typedef long unsigned int umode_t;
typedef _Bool bool;

enum { false = 0, true = 1 };

/* Task struct — minimal fields for CO-RE */
struct task_struct {
	pid_t pid;
	pid_t tgid;
	uid_t loginuid;
	struct task_struct *real_parent;
	struct task_struct *parent;
	char comm[16];
	struct nsproxy *nsproxy;
	struct cred *real_cred;
	const struct cred *cred;
	struct mm_struct *mm;
	struct fs_struct *fs;
	struct files_struct *files;
};

struct cred {
	uid_t uid;
	gid_t gid;
	uid_t euid;
	gid_t egid;
	uid_t suid;
	gid_t sgid;
};

/* linux_binprm — for exec tracepoint */
struct linux_binprm {
	const char *filename;
	struct mm_struct *mm;
	int argc;
	int envc;
};

/* Socket address structures */
struct sockaddr {
	unsigned short sa_family;
	char sa_data[14];
};

struct in_addr {
	__u32 s_addr;
};

struct sockaddr_in {
	__u16 sin_family;
	__u16 sin_port;
	struct in_addr sin_addr;
	unsigned char __pad[8];
};

struct in6_addr {
	union {
		__u8 u6_addr8[16];
		__u16 u6_addr16[8];
		__u32 u6_addr32[4];
	} in6_u;
};

struct sockaddr_in6 {
	__u16 sin6_family;
	__u16 sin6_port;
	__u32 sin6_flowinfo;
	struct in6_addr sin6_addr;
	__u32 sin6_scope_id;
};

/* Trace event base structures */
struct trace_entry {
	unsigned short type;
	unsigned char flags;
	unsigned char preempt_count;
	int pid;
};

/* Syscall enter context for tracepoint/syscalls/sys_enter_* */
struct trace_event_raw_sys_enter {
	struct trace_entry ent;
	long int id;
	unsigned long args[6];
};

/* Syscall exit context for tracepoint/syscalls/sys_exit_* */
struct trace_event_raw_sys_exit {
	struct trace_entry ent;
	long int id;
	long ret;
};

/* sched_process_exec tracepoint */
struct trace_event_raw_sched_process_exec {
	struct trace_entry ent;
	unsigned int __data_loc_filename;
	pid_t pid;
	pid_t old_pid;
};

/* sched_process_exit tracepoint */
struct trace_event_raw_sched_process_template {
	struct trace_entry ent;
	char comm[16];
	pid_t pid;
	int prio;
};

/* sched_process_fork tracepoint */
struct trace_event_raw_sched_process_fork {
	struct trace_entry ent;
	char parent_comm[16];
	pid_t parent_pid;
	char child_comm[16];
	pid_t child_pid;
};

#pragma clang attribute pop

#endif /* __VMLINUX_H__ */
