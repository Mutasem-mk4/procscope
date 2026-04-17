# Acceptance Risk Assessment

Last updated: 2026-04-17

## Status: NOT YET SUBMISSION-READY

procscope is structurally complete and honestly documented, but has external
blockers that prevent it from being submitted to any distro right now.

## What Is Ready

| Area | Status | Details |
|------|--------|---------|
| Source builds | ✅ | `go build`, `go vet`, `go test` pass on Windows, Linux cross-build passes |
| Committed BPF artifact | ✅ | `internal/tracer/procscope_bpfel.o` (68KB) embedded via `go:embed` |
| Debian packaging skeleton | ✅ | `debian/rules` builds from committed artifact, no `go generate` |
| Arch PKGBUILD | ⚠️ | Structurally correct; sha256sums=SKIP (see below) |
| Documentation | ✅ | Honest support matrix, no false claims |
| Man page | ✅ | `man/procscope.1` |
| Shell completions | ✅ | bash, zsh, fish |
| License | ✅ | MIT, DEP-5 copyright |
| Changelog | ✅ | Keep a Changelog format |
| Security policy | ✅ | Threat model, capabilities table |
| Unit tests | ✅ | 28/28 pass |
| Module integrity | ✅ | `go mod verify` passes, `go.sum` committed |
| CI/CD config | ✅ | GitHub Actions + GoReleaser configured |

## External Blockers (Cannot Be Resolved Locally)

### 1. No Public Repository

The GitHub repository `github.com/Mutasem-mk4/procscope` does not exist yet.
Every distro submission requires a publicly accessible upstream repository.

**To resolve:** Create the repository on GitHub, push all commits and the
`v0.1.0` tag.

### 5. PKGBUILD sha256sums=SKIP

The in-tree PKGBUILD uses `sha256sums=('SKIP')` because an in-tree file
cannot contain its own archive's hash (bootstrap problem). This is standard
for upstream convenience PKGBUILDs. The actual BlackArch submission will be
a copy of this PKGBUILD with the real hash filled in.

**To resolve:** After pushing the tag, compute the hash from the GitHub
tarball:
```bash
curl -sL https://github.com/Mutasem-mk4/procscope/archive/v0.1.0.tar.gz | sha256sum
```

### 2. No Linux Runtime Verification

procscope has never run on a real Linux host. The eBPF programs have been
compiled but never loaded into a kernel. There may be:

- BPF verifier rejections
- Struct layout mismatches between the Go types and BPF event struct
- Tracepoint attachment failures on specific kernel versions
- Ring buffer read issues

**To resolve:** Run on a Linux host with kernel 5.8+ and BTF:
```bash
sudo ./bin/procscope -- ls /tmp
```

### 3. No Package Build Verification

Neither `dpkg-buildpackage` nor `makepkg` has been run against this repo.
The packaging files are structurally correct based on documentation and
policy review, but have not been validated by the actual packaging tools.

**To resolve:** On Debian/Kali: `dpkg-buildpackage -us -uc -b` then
`lintian ../procscope_*.deb`. On Arch: `cd arch && makepkg -sf`.

### 4. Placeholder Maintainer Identity

All packaging files use `procscope contributors <security@procscope.dev>`.
This is a placeholder — the domain and email do not exist. Distro
maintainers will want a real person's name and reachable email.

**To resolve:** Replace with a real name and email before submission.

## Distro-Specific Submission Requirements

### Kali Linux

| Requirement | Status |
|------------|--------|
| Public GitHub repo | ❌ Not created |
| Debian package builds | ❌ Not verified |
| Tool demonstrates value | ⚠️ Not runtime-tested |
| Active upstream | ❌ No public commit history |
| Kali bug tracker request | ❌ Not filed |

### BlackArch

| Requirement | Status |
|------------|--------|
| Public GitHub repo | ❌ Not created |
| PKGBUILD with real sha256 | ⚠️ SKIP — fill from GitHub tarball after push |
| PKGBUILD builds cleanly | ❌ Not verified with makepkg |
| Tool in `blackarch-forensic` or `blackarch-debugging` | ⚠️ Planned |
| PR to BlackArch repo | ❌ Not filed |

### Parrot Security

| Requirement | Status |
|------------|--------|
| Debian-standard package | ❌ dpkg-buildpackage not verified |
| Public upstream | ❌ Not created |
| Community contribution submission | ❌ Not filed |

## Risk Summary

| Risk | Severity | Mitigation |
|------|----------|------------|
| BPF verifier rejects programs | HIGH | Test on Linux; fix C code if needed |
| Go struct ↔ BPF struct mismatch | MEDIUM | `binary.Read` will fail loudly; fixable |
| Packaging tools reject files | LOW | Skeleton follows policy closely |
| PKGBUILD sha256 = SKIP | LOW | Fill in after cutting GitHub release |
| Placeholder email rejected | LOW | Simple text replacement |

## Recommended Next Steps (In Priority Order)

1. Create `github.com/Mutasem-mk4/procscope` and push
2. Boot a Kali/Ubuntu VM with kernel 5.8+ and BTF
3. Run `make build && sudo ./bin/procscope -- ls /tmp`
4. Fix any BPF verifier or struct mismatch issues
5. Run `dpkg-buildpackage` and `lintian`
6. Run `makepkg` on Arch
7. Replace placeholder maintainer email
8. File Kali bug tracker tool request
9. Submit BlackArch PKGBUILD PR
