#!/usr/bin/env python3
"""
Release preflight checks for packaging consistency.
"""

from __future__ import annotations

import argparse
import pathlib
import re
import sys


ROOT = pathlib.Path(__file__).resolve().parents[1]


def read_text(path: pathlib.Path) -> str:
    return path.read_text(encoding="utf-8")


def expect_contains(label: str, text: str, needle: str, errors: list[str]) -> None:
    if needle not in text:
        errors.append(f"{label} missing expected content: {needle}")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--tag", required=True, help="Release tag, e.g. v0.1.4")
    args = parser.parse_args()

    tag = args.tag.strip()
    version = tag[1:] if tag.startswith("v") else tag
    errors: list[str] = []

    changelog = read_text(ROOT / "CHANGELOG.md")
    expect_contains("CHANGELOG.md", changelog, f"## [{version}]", errors)

    debian_changelog = read_text(ROOT / "debian" / "changelog")
    expect_contains("debian/changelog", debian_changelog, f"procscope ({version}-", errors)

    pkgbuild = read_text(ROOT / "arch" / "PKGBUILD")
    srcinfo = read_text(ROOT / "arch" / ".SRCINFO")
    expect_contains("arch/PKGBUILD", pkgbuild, f"pkgver={version}", errors)
    expect_contains("arch/.SRCINFO", srcinfo, f"\tpkgver = {version}", errors)

    if errors:
        print("release preflight failed:")
        for err in errors:
            print(f" - {err}")
        return 1

    print(f"release preflight passed for {tag}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
