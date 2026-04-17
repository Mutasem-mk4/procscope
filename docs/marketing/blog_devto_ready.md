---
title: Why I built procscope instead of just using Tracee or Falco
published: true
description: I was tired of host-wide EDR noise during malware triage. So I built an eBPF sniper rifle.
tags: security, ebpf, golang, linux
cover_image: https://raw.githubusercontent.com/Mutasem-mk4/procscope/master/assets/header.png
canonical_url: https://github.com/Mutasem-mk4/procscope
---

If you work in cloud security, you are probably exhausted by the acronym soup of eBPF tools. From Sysdig's Falco to Aqua's Tracee and Isovalent's Tetragon, the market is flooded with "Kernel-Level Observability" daemons.

So why on earth did I write [procscope](https://github.com/Mutasem-mk4/procscope)?

**Because enterprise EDRs are built for fleets, not for incident responders.**

## The Signal-to-Noise Problem

When I get paged because a suspicious container is running a slightly strange binary, I don't want to deploy a massive DaemonSet across my cluster, configure 30 YAML policy files, and then try to grep through 10,000 JSON logs to find the single `/etc/shadow` file access.

I just want to know what the binary is doing *right now*.

If I use `strace -p`, any cleverly written malware instantly detects the `ptrace` system call and goes silent (or worse, bombs the host).

## Enter procscope

`procscope` was designed to be the **tactical sniper rifle** of eBPF, not the shotgun.

It is a single, statically compiled binary. You don't configure policies. You don't deploy an agent. You literally just pass it the command you want to trace like this:

```bash
sudo procscope -- ./sketchy_installer.sh
```

Under the hood, `procscope` injects kprobes and tracepoints into the kernel, but filters events **at the kernel-level** so it *only* evaluates events originating from that specific PID tree.

Tetragon monitors the entire host. `procscope` monitors the blast radius.

## What you see in real time

The timeline spit out on your terminal contains:
- 🌐 Every IP address and port the binary beaconed out to
- 🗂️ Every file it touched, read, or deleted
- 🔐 Every privilege escalation attempt (`setuid`/`chown`)
- ☸️ The Kubernetes Pod and Namespace it belongs to (v1.1.0+)

**Zero noise. Zero host-wide overhead.**

## Try it right now — in your browser

No installation required. We built a [free interactive sandbox on Killercoda](https://killercoda.com/mutasem04/scenario/procscope-scenario) where you can trace a fake reverse-shell payload in under 60 seconds.

[![Try it in the Browser](https://img.shields.io/badge/Try_in_Browser-Killercoda-23C13F?style=for-the-badge&logoColor=white)](https://killercoda.com/mutasem04/scenario/procscope-scenario)

## Getting it

```bash
# Linux amd64 — single binary, no dependencies
curl -sL https://github.com/Mutasem-mk4/procscope/releases/download/v1.1.0/procscope_1.1.0_linux_amd64.tar.gz | tar -xz
sudo mv procscope /usr/local/bin/
```

Star it on GitHub: [github.com/Mutasem-mk4/procscope](https://github.com/Mutasem-mk4/procscope)

---

*If you are an incident responder or malware reverse engineer tired of deploying massive agents just to trace a single shell script — this tool is for you.*
