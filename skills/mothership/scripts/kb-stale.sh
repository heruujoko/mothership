#!/usr/bin/env bash
# kb-stale.sh — Report stale wiki pages
# Usage: bash skills/mothership/scripts/kb-stale.sh
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
wiki_config="${repo_root}/.mothership/wiki.yaml"
helper="${repo_root}/skills/mothership/scripts/wiki-hygiene.py"

if [ ! -f "${wiki_config}" ]; then
  echo "No wiki configured (${wiki_config} not found)."
  exit 0
fi

python3 "${helper}" kb-stale --config "${wiki_config}"
