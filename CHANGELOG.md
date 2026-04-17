# Changelog

All notable changes to procscope will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-04-16

### Added

- Initial release of procscope
- Process lifecycle tracing: exec, fork/clone, exit with codes
- File activity tracing: openat, rename, unlink, chmod, chown
- Network activity tracing: connect, accept, bind, listen with IP:port
- Privilege transition detection: setuid, setgid, ptrace
- Namespace change detection: setns, unshare
- Mount operation detection
- Command launch mode (`procscope -- ./command`)
- PID attach mode (`procscope -p PID`)
- Process name attach mode (`procscope -n name`)
- Automatic fork-following (children auto-tracked)
- Live terminal timeline with ANSI colors
- JSONL event stream output
- Evidence bundle directory with:
  - metadata.json
  - events.jsonl
  - process-tree.txt
  - files.json
  - network.json
  - notable.json
  - summary.md
- Markdown executive summary report
- Runtime capability and privilege detection
- Safe defaults: no env dumping, bounded args/paths, sensitive pattern redaction
- Versioned event schema (v1.0.0)
- Shell completions (bash, zsh, fish)
- Man page
- Debian packaging (debian/)
- Arch/BlackArch packaging (arch/PKGBUILD)
- CI/CD workflows (GitHub Actions)
- Comprehensive documentation
- Unit and integration tests

### Known Limitations

- DNS query extraction is not implemented in v0.1.0 (event type defined in schema for future use)
- File paths from openat may be relative when using dirfd != AT_FDCWD
- Static binaries may not trigger expected syscall probes
- Container ID extraction is not implemented (schema field reserved for future use)
- Event drops possible under very high event rates
- Requires kernel 5.8+ with BTF

[Unreleased]: https://github.com/Mutasem-mk4/procscope/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Mutasem-mk4/procscope/releases/tag/v0.1.0
