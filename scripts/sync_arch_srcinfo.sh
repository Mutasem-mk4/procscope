#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root/arch"

if ! command -v makepkg >/dev/null 2>&1; then
  echo "makepkg is required to regenerate .SRCINFO" >&2
  exit 1
fi

makepkg --printsrcinfo > .SRCINFO
echo "Updated arch/.SRCINFO from arch/PKGBUILD"
