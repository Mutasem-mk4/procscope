# Blog Draft: Why I built procscope instead of just using Tracee or Falco

If you work in cloud security, you are probably exhausted by the acronym soup of eBPF tools. From Sysdig's Falco to Aqua's Tracee and Isovalent's Tetragon, the market is flooded with "Kernel-Level Observability" daemons.

So why on earth did I write [procscope](https://github.com/Mutasem-mk4/procscope)?

Because enterprise EDRs are built for fleets, not for incident responders.

### The Signal-to-Noise Problem
When I get paged because a suspicious container is running slightly strange binary, I don't want to deploy a massive DaemonSet across my cluster, configure 30 YAML policy files, and then try to grep through 10,000 JSON logs to find the single `/etc/shadow` file access. 

I just want to know what the binary is doing *right now*.

If I use `strace -p`, any cleverly written malware instantly detects the `ptrace` system call and goes silent (or worse, bombs the host). 

### Enter procscope
`procscope` was designed to be the tactical sniper rifle of eBPF, not the shotgun. 

It is a single, statically compiled binary. You don't configure policies. You don't deploy an agent. You literally just pass it the command you want to run. 

```bash
sudo procscope -- ./sketchy_installer.sh
```

Under the hood, `procscope` injects kprobes and tracepoints into the kernel, but it filters events at the kernel-level so it *only* evaluates events originating from that specific PID tree. 

Tetragon monitors the entire host. `procscope` monitors the blast radius.

The result? Absolute zero noise. The timeline spit out on your terminal contains exactly every file the binary touched, every IP address it beaconed out to, and every child process it forked, with Kubernetes Namespace resolution stitched right in.

If you are an incident responder or malware reverse engineer tired of deploying massive agents just to trace a single shell script, rip the binary from GitHub here: [github.com/Mutasem-mk4/procscope](https://github.com/Mutasem-mk4/procscope).
