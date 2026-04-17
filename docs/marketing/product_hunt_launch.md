# 🚀 Product Hunt Launch Strategy

**Product Name:** procscope
**Tagline:** See exactly what Linux processes are doing. Uncover malware instantly.
**Topics:** Developer Tools, Cybersecurity, Open Source

### Main Description
Meet `procscope` — the ultimate eBPF-powered runtime investigator.

Whether you are reverse-engineering a suspicious binary, chasing down a 0-day, or debugging a misbehaving CI script, `procscope` lets you attach to any process and instantly see its entire lifecycle in real-time.

Instead of writing complex BPF programs or dealing with the paralyzing noise of host-wide tools like Falco, you just run:
`sudo procscope -- ./suspicious-binary`

What you get:
🌐 Instant visibility of all network connections made (even through curl/nc)
🗂️ File opens, reads, and deletes
🔐 Privilege escalation attempts (setuid/chown)
☸️ Kubernetes Pod/Namespace automatic resolution!

Stop guessing what scripts are doing. See it.

### Maker Comment (Post immediately after launch)
Hey everyone! 👋 Maker here. 

As a cybersecurity engineer, I got sick of relying on `strace` (which freezes malware with ptrace overhead) or trying to parse massive `auditd` logs just to see if a script was making a malicious network call. 

I built `procscope` using eBPF to fix this. It runs with zero runtime overhead, traces *only* the process you tell it to, and spits out a beautiful timeline of exactly what a process is trying to hide from you. I just rolled out native Kubernetes support in v1.1.0 today!

I’d love to hear your feedback on the JSON pipelines or any new syscalls you want hooked! Happy to answer any technical questions.
