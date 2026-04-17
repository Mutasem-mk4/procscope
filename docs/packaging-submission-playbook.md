# Distribution Submission Playbook

This playbook tracks what is required to keep `procscope` eligible for security-focused distributions.
It is a readiness checklist, not a guarantee of acceptance.

## BlackArch

Target path: AUR/BlackArch packaging based on `arch/PKGBUILD`.

Checklist:

- Keep `arch/PKGBUILD` and `arch/.SRCINFO` synchronized for every release.
- Ensure `pkgver`, source URL, and checksum match the release tarball.
- Keep build reproducible (`CGO_ENABLED=0`, trimmed paths, deterministic source fetch).
- Include man page, shell completions, and license in package install stage.
- Run `namcap` and review warnings before submission updates.
- Validate on a real Arch or BlackArch host before opening the PR.

## Kali Linux

Target path: Debian package sponsorship and eventual inclusion.

Checklist:

- Maintain `debian/control`, `debian/changelog`, `debian/rules`, and tests.
- Keep `Standards-Version` current and remove avoidable lintian warnings.
- Ensure package metadata reflects stable ABI and runtime requirements.
- Provide a concise threat-model and incident-response use case in package description.
- Keep changelog entries specific and release-aligned.
- Build and test on a Debian-family Linux host before opening maintainer outreach.

## Parrot OS

Parrot typically follows Debian packaging patterns and may consume upstream Debian work.

Checklist:

- Keep Debian packaging metadata clean and policy-compliant.
- Keep dependency surface minimal and explicit.
- Maintain reproducible binaries and complete source availability.
- Keep release notes security-focused and operationally useful.
- Treat Debian package quality as the primary submission gate.

## Release Gate (Must Pass)

Before tag/release:

- CI green (`.github/workflows/ci.yml`)
- Packaging quality workflow green (`.github/workflows/packaging-quality.yml`)
- Arch metadata synced
- Debian metadata validated
- Linux smoke test complete
- Release notes include packaging-impact section

## Maintainer Outreach Notes

When opening distro requests:

- Lead with process-scoped security triage use case.
- Link to stable release, checksums, and signed tags if available.
- Include evidence of CI + packaging validation.
- Provide a quick smoke test command and expected output.
- Avoid promising distro availability until the package is actually merged.
