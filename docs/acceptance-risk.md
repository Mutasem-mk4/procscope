# Acceptance Risk Assessment

Last updated: 2026-04-17

## Status: CLOSE, BUT NOT YET SUBMISSION-READY

`procscope` already has the basic upstream signals distro maintainers expect:

- public source repository
- tagged releases
- in-tree Debian packaging
- in-tree Arch packaging
- man page, shell completions, and license metadata
- automated CI and unit tests

The remaining blockers are mostly about Linux-side verification and release discipline, not missing project structure.

## Verified Locally on 2026-04-17

The current `master` branch was checked from a Windows development host:

- `go test ./...` passes
- `go vet ./...` passes
- repo state is clean relative to `origin/master`

This is useful, but it is not enough for distro submission because neither package builds nor runtime eBPF behavior were validated on Linux in this audit.

## What Is Already Strong

| Area | Status | Notes |
|------|--------|-------|
| Public upstream | ✅ | GitHub repo and release tags exist |
| Debian metadata | ✅ | `debian/control`, `debian/rules`, `debian/watch`, DEP-8 tests present |
| Arch metadata | ✅ | `arch/PKGBUILD` and `arch/.SRCINFO` present |
| Build reproducibility direction | ✅ | committed BPF object, `CGO_ENABLED=0`, `-trimpath` |
| User docs | ✅ | README, BUILDING, packaging, support matrix, architecture docs |
| Maintainer docs | ✅ | submission playbook, security policy, contribution guide |
| Quality gates | ✅ | CI, packaging workflow, release preflight script |

## Remaining Blockers

### 1. Linux Runtime Validation Is Still Missing

The most important unresolved risk is real-kernel behavior on the target distros:

- eBPF verifier acceptance
- tracepoint availability across distro kernels
- capability handling on Kali / Parrot / Arch-family systems
- end-to-end launch mode and attach mode behavior

Minimum validation to complete before submission:

```bash
make build
sudo ./bin/procscope -- /bin/true
sudo ./bin/procscope -p <existing-pid>
sudo ./bin/procscope --out case-001 --summary report.md -- /bin/ls /tmp
```

### 2. Native Package Tooling Has Not Been Re-Validated on Linux

Package metadata exists, but it still needs actual Linux-side verification:

```bash
# Debian / Kali / Parrot
dpkg-buildpackage -us -uc -b
lintian ../procscope_*.changes
autopkgtest . -- null

# Arch / BlackArch
cd arch
makepkg -sf
namcap PKGBUILD
namcap ./*.pkg.tar.zst
```

### 3. Release Metadata Must Stay Aligned Before the Next Tag

This repo had version drift across tags, package metadata, and README examples.
That kind of inconsistency slows down maintainer review quickly.

Before the next release, confirm all of these point to the same version:

- `CHANGELOG.md`
- `debian/changelog`
- `arch/PKGBUILD`
- `arch/.SRCINFO`
- README release references

### 4. A Reachable Maintainer Address Is Still Recommended

The current package metadata uses a GitHub noreply address. That is acceptable for source hosting, but distro maintainers often prefer a stable contact address for packaging follow-up.

### 5. Submission Still Depends on Human Maintainer Review

Even a technically clean package can be declined or delayed if:

- the tool category is unclear
- the use case overlaps heavily with existing packages
- upstream appears inactive
- there is no maintainer responsiveness after submission

## Distro-Specific Readiness

| Distro | Current Readiness | Main Gap |
|--------|-------------------|----------|
| Kali Linux | Near-ready | needs Debian package validation on Linux and submission request |
| Parrot OS | Near-ready | needs Debian package validation on Linux and maintainer outreach |
| BlackArch | Near-ready | needs Linux `makepkg` / `namcap` validation and PR submission |

## Recommended Next Steps

1. Validate runtime behavior on a real Kali or Debian host and on an Arch-based host.
2. Build and lint the Debian and Arch packages on Linux, not Windows.
3. Cut the next release only after package metadata and changelogs are aligned.
4. Use the submission playbook in [`docs/packaging-submission-playbook.md`](packaging-submission-playbook.md).
5. Open distro requests with concise evidence: release tag, checksums, CI status, package build logs, and one reproducible smoke test.
