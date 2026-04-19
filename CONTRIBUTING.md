# Contributing to procscope

Thank you for your interest in improving `procscope`! As a security-focused project, we value high-quality contributions that maintain the integrity and performance of the tool.

## Getting Started

1. **Fork the repository** on GitHub.
2. **Clone your fork** locally: `git clone https://github.com/YOUR-USERNAME/procscope.git`
3. **Create a new branch** for your feature or fix: `git checkout -b feature/your-feature-name`
4. **Install dependencies**: Ensure you have Go 1.24+ and the BPF toolchain (`clang`, `llvm`, `libbpf-dev`) installed.

## Development Workflow

### Building from Source
```bash
make build
```

### Running Tests
All contributions must pass existing tests and include new tests where applicable.
```bash
# Run unit tests
go test -v ./...

# Run integration tests (requires root/CAP_SYS_ADMIN)
sudo ./bin/procscope -- ls /tmp
```

### Code Style
We use `golangci-lint` to maintain code quality. Please run the linter before submitting a PR:
```bash
golangci-lint run
```

## Pull Request Process

1. Ensure your code follows the existing style and architectural patterns.
2. **Sign your commits**: We require signed commits to ensure provenance and security.
3. Update the `README.md` or documentation if you're adding new features.
4. Submit the PR against the `master` branch.
5. At least one maintainer review is required before merging.

## Security Contributions

If you find a security vulnerability, please do **NOT** open a public issue. Follow the instructions in [SECURITY.md](SECURITY.md) to report it privately.

## Community & Governance

`procscope` is an open-source project maintained by [Mutasem Kharma](https://github.com/Mutasem-mk4). We are committed to a transparent and welcoming environment for all contributors.
