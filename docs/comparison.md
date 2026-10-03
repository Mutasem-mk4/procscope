# Choosing procscope

Use procscope when the investigation scope is one command or an existing PID,
including its observed descendants, and you want a timeline, JSONL evidence,
process tree and Markdown summary from a short CLI session.

## What the workflow provides

- Command and PID attachment without a persistent monitoring service.
- File, process and network-connect events selected by the tracer.
- A private evidence directory and reports that can be read and exported.
- JSONL for further analysis, with traced-command output kept off the JSON stream.

## What to evaluate before choosing it

procscope does not record every syscall. Network-connect events are attempts,
not proof of completed connections. It requires the documented Linux kernel,
BTF and tracing privileges. Container context is limited; it is not a complete
container security policy or enforcement system.

Tracing has overhead and can lose events. This repository does not provide a
controlled comparative benchmark against strace, Tracee, Tetragon or sysdig;
therefore it makes no relative performance, capture-completeness or deployment-size
claim. A successful fixture trace does not establish those properties.

## Other workflows

If you need syscall-level debugging, system-wide capture, Kubernetes integration
or enforcement, evaluate tools designed for that scope. Consult their current
documentation rather than treating this project's comparison as a feature audit:

- [strace](https://strace.io/)
- [Tracee](https://aquasecurity.github.io/tracee/)
- [Tetragon](https://tetragon.io/docs/)
- [Inspektor Gadget](https://www.inspektor-gadget.io/docs/)
- [sysdig](https://github.com/draios/sysdig)

The reason to choose procscope is its focused investigation and reporting workflow.
Choose a broader tool when that workflow does not meet the investigation needs.
