# 🎤 Conference Call For Papers (CFP) Draft
*Targeted for DEF CON Demo Labs or BlackHat Arsenal*

**Title:** "Micro-Tracing Malware: Dropping EDR Overhead for Targeted eBPF Investigations"
**Author:** Mutasem Kharma

### Abstract
When incident responders or reverse engineers discover an obfuscated shell script or a suspicious binary on a Linux host, standard telemetry usually fails them. `strace` alerts evasion mechanisms via `ptrace` detection, and host-wide EDRs like Falco create thousands of noise events every second. 

In this Demo Lab, I will introduce `procscope`—an open-source, eBPF-powered process runtime investigator. Rather than writing complex BPF hooks or managing daemon policies, users simply prefix `procscope` to any command. Doing so dynamically instruments the kernel to track that specific process and its children in real-time, instantly logging network beacons, file access, and namespace escapes without triggering anti-debugging traps.

### Speaker Outline
1. **The Problem (2 mins):** Why modern malware easily evades `strace` and how host-wide security daemons create paralyzing signal-to-noise ratios.
2. **The eBPF Solution (5 mins):** Walking through the architecture of `procscope`. How it safely attaches `kprobes` and uses bounded ring-buffers to maintain 0% CPU overhead while observing `execve`, `connect`, and `setuid` events.
3. **Live Arsenal Demo (5 mins):** Launching a highly obfuscated script. I will demonstrate how `procscope` instantly catches the script downloading a secondary payload and pivoting to a reverse shell, complete with automatic Kubernetes Pod resolution.
4. **Q&A (3 mins)**

### Key Takeaway
Attendees will walk away with an open-source, statically compiled tool they can deploy to any modern Linux server to instantly triage misbehaving scripts and active malware without configuring a massive SIEM infrastructure.
