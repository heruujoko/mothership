#!/usr/bin/env bash

set -euo pipefail

repo_root="$(
  cd "$(dirname "${BASH_SOURCE[0]}")" && pwd
)"
installer_dir="$repo_root/tools/installer"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required to run the installer." >&2
  exit 1
fi

cd "$installer_dir"
exec go run . --repo-root "$repo_root" "$@"
