# Packaging Guide

## Debian / Kali / Parrot

### Building the Package

```bash
# Install build dependencies
sudo apt install debhelper golang-go

# Build the package
dpkg-buildpackage -us -uc -b

# Install
sudo dpkg -i ../procscope_*.deb
```

### Package Layout

```
debian/
├── control          # Package metadata, deps, description
├── rules            # Build rules (dh + go build)
├── changelog        # Package changelog
├── copyright        # Machine-readable copyright (DEP-5)
├── watch            # Upstream version tracking
├── source/format    # 3.0 (quilt)
└── tests/
    ├── control      # DEP-8 autopkgtest definitions
    ├── help-test    # Verify --help works
    └── smoke-test   # Verify --version works
```

The Debian package builds from the committed `internal/tracer/procscope_bpfel.o`
artifact and does not run `go generate` during package build.

### Kali Tool Submission

The repository is structured to support a Kali package request and Debian-style sponsorship work. Keep the request grounded in current packaging evidence, not aspirational install commands.

| Required Field | Value |
|---------------|-------|
| Name | procscope |
| Homepage | https://github.com/Mutasem-mk4/procscope |
| License | MIT |
| Description | Process-scoped runtime investigation tool using eBPF |
| Similar tools | strace, ltrace, sysdig |
| Activity | Active development |
| Current install path | GitHub release asset or `go install` |
| Usage | `sudo procscope -- ./binary` |

### Parrot Contribution

Parrot follows a Debian-oriented packaging flow closely enough that the Debian package quality is the main gate. Finish Debian package validation first, then open maintainer outreach with package build logs and a short smoke test transcript.

## Arch Linux / BlackArch

### Building the Package

```bash
cd arch/
makepkg -si
```

### PKGBUILD Notes

- Follows Arch Go packaging guidelines
- Consumes the committed `internal/tracer/procscope_bpfel.o` artifact instead of invoking `go generate`
- No network access during build (offline build)
- Respects system build flags
- Installs to standard paths (`/usr/bin`, `/usr/share/man`, `/usr/share/licenses`)

### BlackArch Submission

Prepare the PKGBUILD as if it will be reviewed by Arch maintainers first:

1. Fork the BlackArch repository
2. Regenerate `arch/.SRCINFO`
3. Run `makepkg` and `namcap`
4. Add the package to the appropriate category (`blackarch-forensic` or `blackarch-debugging`)
5. Submit pull request

## Release Process

### Version Tagging

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

### GoReleaser

```bash
goreleaser release --clean
```

This creates:
- Linux binaries (amd64, arm64)
- Checksums
- Release notes from CHANGELOG.md

### Manual Release Checklist

1. Update `CHANGELOG.md`
2. Run release preflight checks: `python scripts/release_preflight.py --tag vX.Y.Z`
3. Sync Arch metadata when `arch/PKGBUILD` changes: `./scripts/sync_arch_srcinfo.sh`
4. Build and smoke test on Linux: `make build && sudo ./bin/procscope -- /bin/true`
5. Build Debian package: `dpkg-buildpackage -us -uc -b`
6. Build Arch package: `cd arch && makepkg -sf`
7. Tag the release
8. Run `goreleaser release`
9. Upload artifacts to GitHub release
10. Update documentation if needed
