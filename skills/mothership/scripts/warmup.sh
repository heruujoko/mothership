#!/usr/bin/env bash
set -euo pipefail

repo_root="$(pwd)"
runtime_dir="${repo_root}/.mothership"
preflight_dir="${runtime_dir}/preflight"
hub_dir="${runtime_dir}/hub"
hub_checkpoints_dir="${hub_dir}/checkpoints"
report_file="${preflight_dir}/skills-status.md"
dependencies_file="${repo_root}/mothership-config/skill-dependencies.md"
gitignore_file="${repo_root}/.gitignore"

mkdir -p "${runtime_dir}" "${preflight_dir}" "${runtime_dir}/sessions" "${hub_checkpoints_dir}"

if [ ! -f "${gitignore_file}" ]; then
  touch "${gitignore_file}"
fi

if ! rg -n '^\.mothership/$' "${gitignore_file}" >/dev/null 2>&1; then
  printf "\n.mothership/\n" >> "${gitignore_file}"
fi

timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

required_skills=(
  "obra/superpowers"
  "obra/creating-plan"
  "obra/systematical-debugging"
)

codex_home="${CODEX_HOME:-${HOME}/.codex}"
search_roots=(
  "${repo_root}/skills"
  "${codex_home}/skills"
)

{
  echo "# Mothership Skill Preflight"
  echo
  echo "- generated_at: ${timestamp}"
  echo "- dependencies_source: \`mothership-config/skill-dependencies.md\`"
  echo
  echo "## Required Skills"
} > "${report_file}"

missing_count=0
for skill in "${required_skills[@]}"; do
  found_path=""
  for root in "${search_roots[@]}"; do
    candidate="${root}/${skill}/SKILL.md"
    if [ -f "${candidate}" ]; then
      found_path="${candidate}"
      break
    fi
  done

  if [ -n "${found_path}" ]; then
    echo "- [ok] \`${skill}\` (${found_path})" >> "${report_file}"
  else
    echo "- [missing] \`${skill}\`" >> "${report_file}"
    missing_count=$((missing_count + 1))
  fi
done

{
  echo
  echo "## Runtime Hygiene"
  echo "- runtime_dir: \`${runtime_dir}\`"
  echo "- gitignore_rule: \`.mothership/\`"
  echo "- dependencies_file_present: $([ -f "${dependencies_file}" ] && echo "yes" || echo "no")"
} >> "${report_file}"

if [ "${missing_count}" -gt 0 ]; then
  echo "warmup: completed with missing skills (${missing_count}). see ${report_file}" >&2
  exit 2
fi

echo "warmup: completed. report: ${report_file}"
