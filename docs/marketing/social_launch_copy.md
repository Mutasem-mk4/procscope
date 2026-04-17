# 📣 Social Launch Copy — procscope

## HackerNews (Show HN)

**Title:**
`Show HN: procscope – eBPF process tracer for malware triage without strace overhead`

**Body to paste in comments:**
```
Hey HN! I built procscope because I got tired of choosing between the ptrace overhead 
of strace (which breaks timing-sensitive malware) and deploying a full EDR like Falco 
just to figure out what one suspicious binary is doing.

It's a single static binary. You prefix it to any command and get a clean timeline of 
every file, network connection, and privilege escalation that process tree triggers — 
using eBPF at the kernel level with zero overhead.

v1.1.0 now resolves Kubernetes Pod/Namespace metadata automatically.

You can try it in your browser without installing anything:
https://killercoda.com/mutasem04/scenario/procscope-scenario

GitHub: https://github.com/Mutasem-mk4/procscope
Happy to answer questions about the eBPF internals or the K8s resolution approach!
```

**Submit at:** https://news.ycombinator.com/submit

---

## LinkedIn Post

I just spent 6 months building the tool I always wished existed during incident response.

Meet **procscope** — a single-binary eBPF process tracer for Linux.

Instead of sifting through 10,000 Falco events or triggering anti-debug traps with `strace`, you just run:

```bash
sudo procscope -- ./suspicious-binary
```

And instantly see every network call, file access, and privilege escalation attempt in a clean timeline.

v1.1.0 even resolves **Kubernetes Pod and Namespace metadata** so you know exactly which workload is misbehaving.

🎮 **Try it live in your browser — no install needed:**
👉 https://killercoda.com/mutasem04/scenario/procscope-scenario

⭐ GitHub: https://github.com/Mutasem-mk4/procscope

#ebpf #cybersecurity #opensource #golang #linux #sre #devsecops #incidentresponse

---

## Twitter/X Thread

**Tweet 1:**
I got tired of choosing between:
→ strace (breaks timing-sensitive malware via ptrace detection)
→ Falco (10k events/sec of noise just to trace 1 binary)

So I built procscope — an eBPF sniper rifle for process triage 🧵

**Tweet 2:**
One command. One binary. Zero config.
`sudo procscope -- ./suspicious-malware.sh`

You instantly see its network beacons, file reads, and privilege escalation attempts.
No host-wide agent. No policy files.

**Tweet 3:**
v1.1.0 just shipped with automatic Kubernetes Pod resolution.
If you're running workloads in K8s, it maps the container ID → Pod name → Namespace automatically.

**Tweet 4:**
Try it live in your browser right now — FREE, no install:
🔗 https://killercoda.com/mutasem04/scenario/procscope-scenario

Star it on GitHub:
⭐ https://github.com/Mutasem-mk4/procscope

#eBPF #Linux #CyberSecurity #OpenSource
