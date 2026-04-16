# Contributing to procscope

Thank you for your interest in contributing to procscope!

## Getting Started

### Prerequisites

- Go 1.22+
- Linux (for eBPF development and testing)
- clang and llvm-strip (for eBPF compilation)
- Make

### Development Setup

```bash
git clone https://github.com/procscope/procscope.git
cd procscope

# Generate eBPF bindings (Linux only)
make generate

# Build
make build

# Run tests
make test

# Run linters
make lint
```

## How to Contribute

### Bug Reports

Open a GitHub issue with:
- procscope version (`procscope --version`)
- Linux kernel version (`uname -r`)
- Distribution name and version
- Exact steps to reproduce
- Expected vs actual behavior
- Relevant output or logs

### Feature Requests

Open a GitHub issue describing:
- The use case / problem
- Proposed solution
- Why existing features don't cover it

### Code Contributions

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass: `make test`
6. Ensure linting is clean: `make lint`
7. Commit with clear messages
8. Open a pull request

### Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- Keep functions focused and testable
- Document exported types and functions
- Use meaningful variable names
- eBPF C code follows kernel coding style

### Commit Messages

Follow conventional commit format:
```
type: short description

Longer explanation if needed.

Fixes #123
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `ci`, `build`, `perf`

## Architecture

See [docs/architecture.md](docs/architecture.md) for codebase structure.

Key packages:
- `internal/tracer/` — eBPF program management
- `internal/events/` — Event types and correlation
- `internal/process/` — Process tree and launching
- `internal/output/` — Rendering (timeline, JSON, bundle, summary)
- `internal/caps/` — Capability detection
- `internal/redact/` — Redaction controls
- `bpf/` — eBPF C programs

## Testing

```bash
# Unit tests
make test-unit

# Integration tests (requires root, Linux)
make test-integration

# Coverage report
make test-cover
```

### Writing Tests

- Unit tests go in `*_test.go` alongside the code
- Integration tests go in `test/integration/`
- Test fixtures (C programs) go in `test/fixtures/`
- Tests must not require network access
- Tests must be deterministic

## Release Process

See [docs/packaging.md](docs/packaging.md) for packaging and release details.

## Code of Conduct

Be respectful and constructive. We're building tools for the security community.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
