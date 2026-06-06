#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/../../.." && pwd)"
helper="${repo_root}/skills/mothership/scripts/wiki-hygiene.py"

echo "[test] warmup creates missing hub state and records stale pages"
tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

mkdir -p "${tmpdir}/wiki/projects/demo/insights"
cat > "${tmpdir}/config.yaml" <<EOF
wiki:
  root: "${tmpdir}/wiki"
  project: demo
  stale_threshold_days: 1
EOF
cat > "${tmpdir}/wiki/projects/demo/insights/old.md" <<'EOF'
---
freshness: 2000-01-01
---
old
EOF
: > "${tmpdir}/report.md"
python3 "${helper}" warmup \
  --config "${tmpdir}/config.yaml" \
  --report-file "${tmpdir}/report.md" \
  --hub-state "${tmpdir}/hub/state.yaml"

grep -Fqx 'wiki:' "${tmpdir}/hub/state.yaml"
grep -Fqx '  stale_pages:' "${tmpdir}/hub/state.yaml"
grep -Fqx '    - projects/demo/insights/old.md' "${tmpdir}/hub/state.yaml"

echo "wiki hygiene tests: PASS"
