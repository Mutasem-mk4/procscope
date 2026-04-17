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

The repository is structured for a Kali tool request per https://www.kali.org/docs/tools/submitting-tools/:

| Required Field | Value |
|---------------|-------|
| Name | procscope |
| Homepage | https://github.com/Mutasem-mk4/procscope |
| License | MIT |
| Description | Process-scoped runtime investigation tool using eBPF |
| Similar tools | strace, ltrace, sysdig |
| Activity | Active development |
| Install | `sudo apt install procscope` |
| Usage | `sudo procscope -- ./binary` |

### Parrot Contribution

Per https://parrotsec.org/docs/introduction/community-contributions/, the tool should be packaged to Debian standards first, then submitted via their contribution process.

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

Per https://github.com/BlackArch/blackarch, submit a PKGBUILD following:
1. Fork the BlackArch repository
2. Add PKGBUILD to the appropriate category (`blackarch-forensic` or `blackarch-debugging`)
3. Submit pull request

## Release Process

### Version Tagging

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
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
4. Tag the release
5. Run `goreleaser release`
6. Build Debian package: `dpkg-buildpackage -us -uc`
7. Build Arch package: `cd arch && makepkg -sf`
8. Upload artifacts to GitHub release
9. Update documentation if needed
