#!/usr/bin/env bash
# kb-stale.sh — Report stale wiki pages
# Usage: bash skills/mothership/scripts/kb-stale.sh
# Reads wiki config from .mothership/wiki.yaml, scans all insight/decision
# pages, and lists those where `freshness` exceeds stale_threshold_days.
set -euo pipefail

repo_root="$(pwd)"
wiki_config="${repo_root}/.mothership/wiki.yaml"

if [ ! -f "${wiki_config}" ]; then
  echo "No wiki configured (${wiki_config} not found)."
  exit 0
fi

# Parse YAML safely with python3
wiki_meta=$(python3 -c "
import yaml, sys
try:
    with open('${wiki_config}') as f:
        cfg = yaml.safe_load(f)
    w = cfg.get('wiki', {})
    root = w.get('root', '')
    project = w.get('project', '')
    stale_days = w.get('stale_threshold_days', 90)
    archive_days = w.get('archive_after_days', 120)
    print(f'root={root}')
    print(f'project={project}')
    print(f'stale_days={stale_days}')
    print(f'archive_days={archive_days}')
except Exception as e:
    print(f'error={e}', file=sys.stderr)
    sys.exit(1)
" 2>/dev/null) || {
  echo "ERROR: Could not parse ${wiki_config}"
  exit 1
}

eval "${wiki_meta}" 2>/dev/null || true

if [ -z "${root}" ] || [ -z "${project}" ]; then
  echo "ERROR: wiki.root or wiki.project not set in ${wiki_config}"
  exit 1
fi

project_dir="${root}/projects/${project}"
stale_threshold="${stale_days:-90}"
archive_threshold="${archive_days:-120}"

if [ ! -d "${project_dir}/insights" ] && [ ! -d "${project_dir}/decisions" ]; then
  echo "No wiki pages found under ${project_dir}"
  exit 0
fi

echo "=== Stale Wiki Pages ==="
echo "Project: ${project}"
echo "Root: ${root}"
echo "Stale threshold: ${stale_threshold} days"
echo "Archive threshold: ${archive_threshold} days"
echo ""

stale_count=0
archive_count=0
now_epoch=$(date +%s)

check_page() {
  local page="$1"
  local freshness freshness_epoch days_old
  freshness=$(grep -E '^freshness:' "${page}" | head -1 | sed 's/^freshness:[[:space:]]*//' | tr -d '\r' || true)
  if [ -z "${freshness}" ]; then
    return  # no freshness field — skip
  fi
  freshness_epoch=$(date -d "${freshness}" +%s 2>/dev/null || echo 0)
  if [ "${freshness_epoch}" -eq 0 ]; then
    return  # unparseable date — skip
  fi
  days_old=$(( (now_epoch - freshness_epoch) / 86400 ))

  local rel="${page#${root}/}"
  if [ "${days_old}" -gt "${archive_threshold}" ] 2>/dev/null; then
    echo "[ARCHIVE] ${rel} — last updated ${freshness} (${days_old} days ago, exceeds ${archive_threshold}d)"
    archive_count=$((archive_count + 1))
  elif [ "${days_old}" -gt "${stale_threshold}" ] 2>/dev/null; then
    echo "[STALE]   ${rel} — last updated ${freshness} (${days_old} days ago)"
    stale_count=$((stale_count + 1))
  fi
}

shopt -s nullglob
for page in "${project_dir}/insights/"*.md "${project_dir}/decisions/"*.md; do
  check_page "${page}"
done
shopt -u nullglob

echo ""
echo "Summary: ${stale_count} stale, ${archive_count} ready for archival"
echo ""
echo "To archive stale pages:"
echo "  mv ${root}/archive/ <page-path>"
echo "  # Then update index.md"