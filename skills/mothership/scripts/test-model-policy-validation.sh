#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/../../.." && pwd)"
validator="${repo_root}/skills/mothership/scripts/validate-model-policy.sh"

echo "[test] valid policy shape"
ALLOW_MISSING_POLICY=0 WARN_ONLY_POLICY=0 "${validator}" "${repo_root}" >/dev/null

echo "[test] fallback when missing policy"
tmp_dir="$(mktemp -d)"
mkdir -p "${tmp_dir}/mothership-config"
ALLOW_MISSING_POLICY=1 WARN_ONLY_POLICY=1 "${validator}" "${tmp_dir}" >/dev/null
rm -rf "${tmp_dir}"

echo "model-policy validation tests: PASS"
