# 🕵️ Tracing a Reverse Shell

First, let's verify that `procscope` is installed.

```bash
procscope --version
```{{exec}}

We have a suspicious file named `payload.sh` in our directory. Let's trace exactly what it does using `procscope`.

Run the following command. The `--` tells procscope that everything after it is the command we want to trace.

```bash
sudo procscope -- ./payload.sh
```{{exec}}

### What just happened?
Notice the clean, colored timeline output? `procscope` used eBPF (Extended Berkeley Packet Filter) to attach to the kernel and watch *only* the syscalls triggered by `payload.sh` and its children.

1. You saw it `exec` a shell.
2. You saw it open the `/etc/passwd` file (Reconnaissance).
3. Most importantly, you saw a loud red `NET` event connecting to `10.0.0.5:4444`. 

You just proved this script is a reverse shell beacon without having to deobfuscate the script or dig through host-wide auditd logs. This is the power of targeted tracing.

When you're done, press `Ctrl+C` to cleanly detach the eBPF probes.
