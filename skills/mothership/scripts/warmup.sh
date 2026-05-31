#!/usr/bin/env bash
set -euo pipefail

repo_root="$(pwd)"
runtime_dir="${repo_root}/.mothership"
preflight_dir="${runtime_dir}/preflight"
hub_dir="${runtime_dir}/hub"
report_file="${preflight_dir}/skills-status.md"
agents_report_file="${preflight_dir}/agents-status.md"
dependencies_file="${repo_root}/mothership-config/skill-dependencies.md"
gitignore_file="${repo_root}/.gitignore"
registry_file="${repo_root}/agents/registry.yaml"
subagent_protocol_file="${repo_root}/skills/mothership/subagent-protocol.md"
model_policy_file="${repo_root}/mothership-config/model-policy.yaml"
model_policy_validator="${repo_root}/skills/mothership/scripts/validate-model-policy.sh"

PI_HOME="${PI_HOME:-${HOME}/.agents}"

mkdir -p "${runtime_dir}" "${preflight_dir}" "${runtime_dir}/sessions"

if [ ! -f "${gitignore_file}" ]; then
  touch "${gitignore_file}"
fi

if ! grep -q '^\.mothership/$' "${gitignore_file}" 2>/dev/null; then
  printf "\n.mothership/\n" >> "${gitignore_file}"
fi

timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

required_skills=(
  "obra/superpowers"
  "obra/brainstorming"
  "obra/making-plans"
  "obra/executing-plans"
  "obra/systematical-debugging"
)

search_roots=(
  "${repo_root}/skills"
  "${CODEX_HOME:-${HOME}/.codex}/skills"
  "${HOME}/.claude/skills"
  "${HOME}/.antigravity/skills"
)

if [ -n "${MOTHERSHIP_SKILLS_HOME:-}" ]; then
  search_roots=(
    "${search_roots[@]:0:1}"
    "${MOTHERSHIP_SKILLS_HOME}"
    "${search_roots[@]:1}"
  )
fi

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

missing_contract_count=0
if [ ! -f "${registry_file}" ]; then
  missing_contract_count=$((missing_contract_count + 1))
fi
if [ ! -f "${subagent_protocol_file}" ]; then
  missing_contract_count=$((missing_contract_count + 1))
fi

{
  echo
  echo "## Host Delegation Files"
  echo "- agent_registry_present: $([ -f "${registry_file}" ] && echo "yes" || echo "no")"
  echo "- subagent_protocol_present: $([ -f "${subagent_protocol_file}" ] && echo "yes" || echo "no")"
  echo
  echo "## Search Roots"
  for root in "${search_roots[@]}"; do
    if [ -n "${root}" ]; then
      echo "- \`${root}\`"
    fi
  done
  echo
  echo "## Runtime Hygiene"
  echo "- runtime_dir: \`${runtime_dir}\`"
  echo "- gitignore_rule: \`.mothership/\`"
  echo "- dependencies_file_present: $([ -f "${dependencies_file}" ] && echo "yes" || echo "no")"
} >> "${report_file}"

{
  echo
  echo "## Model Policy"
  if [ ! -f "${model_policy_file}" ]; then
    echo "- status: missing"
    echo "- fallback: legacy inheritance (current thread model)"
  elif [ ! -x "${model_policy_validator}" ]; then
    echo "- status: validator_missing_or_not_executable"
    echo "- fallback: legacy inheritance (current thread model)"
  else
    set +e
    validator_output="$(ALLOW_MISSING_POLICY=1 WARN_ONLY_POLICY=1 "${model_policy_validator}" "${repo_root}" 2>&1)"
    validator_status=$?
    set -e
    echo "- validator_exit: ${validator_status}"
    echo "- validator_output: |"
    while IFS= read -r line; do
      echo "  ${line}"
    done <<< "${validator_output}"
  fi
} >> "${report_file}"

# --- Design Gap 1: Agent contracts preflight check ---
agent_contracts_dir="${repo_root}/agents"
user_agent_contracts_dir="${PI_HOME}/agents"

# Extract role keys from registry (lines indented with exactly 2 spaces, ending with colon)
registry_roles=""
if [ -f "${registry_file}" ]; then
  registry_roles=$(grep -E '^  [a-z_]+:' "${registry_file}" | sed 's/^  //;s/:.*//' | tr '\n' ' ')
fi

agent_missing_count=0

# Write agents-status report
{
  echo "# Mothership Agent Contracts Preflight"
  echo
  echo "- generated_at: ${timestamp}"
  echo "- registry_file: \`${registry_file}\`"
  echo "- repo_contracts_dir: \`${agent_contracts_dir}\`"
  echo "- user_scope_contracts_dir: \`${user_agent_contracts_dir}\`"
  echo
  echo "## Per-Role Contract Status"
} > "${agents_report_file}"

if [ -z "${registry_roles}" ]; then
  echo "- registry: no roles found (registry missing or empty)" >> "${agents_report_file}"
  agent_missing_count=1
else
  for role in ${registry_roles}; do
    # Skip commander — non-delegated
    if [ "${role}" = "commander" ]; then
      echo "- \`${role}\`: non-delegated role (no contract needed)" >> "${agents_report_file}"
      continue
    fi

    # Honor registry contract_file if specified (e.g., agents/pr_monkey.md)
    contract_name="${role}.md"
    if [ -f "${registry_file}" ]; then
      # Extract contract_file for this role from registry using awk section-scoped extraction
      registry_cf=$(awk -v role="${role}" '
        /^roles:/{ in_roles=1; next }
        /^[a-z_]+:/{ in_roles=0 }
        in_roles && $0 ~ "^  " role ":" { in_role=1; next }
        in_role && $0 ~ "^    contract_file:" {
          sub(/^ *contract_file: */, ""); gsub(/["'"'"']/, ""); print; exit
        }
        in_role && $0 ~ "^  [a-z_]+:" { in_role=0 }
      ' "${registry_file}")
      if [ -n "${registry_cf}" ]; then
        contract_name="$(basename "${registry_cf}")"
      fi
    fi
    repo_contract="${agent_contracts_dir}/${contract_name}"
    user_contract="${user_agent_contracts_dir}/${contract_name}"

    if [ -f "${repo_contract}" ]; then
      echo "- \`${role}\`: ok (repo-local: ${repo_contract})" >> "${agents_report_file}"
    elif [ -f "${user_contract}" ]; then
      echo "- \`${role}\`: ok (user-scope: ${user_contract})" >> "${agents_report_file}"
    else
      echo "- \`${role}\`: MISSING — tried repo-local (\`${repo_contract}\`) and user-scope (\`${user_contract}\`)" >> "${agents_report_file}"
      agent_missing_count=$((agent_missing_count + 1))
    fi
  done

  # Summary line
  echo "" >> "${agents_report_file}"
  if [ "${agent_missing_count}" -eq 0 ]; then
    echo "**Result:** All agent contracts accessible." >> "${agents_report_file}"
  else
    echo "**Result:** ${agent_missing_count} contract(s) missing. Run \`./install.sh --target pi\` to install user-scope contracts." >> "${agents_report_file}"
  fi
fi

missing_contract_count=$((missing_contract_count + agent_missing_count))
# --- End agent contracts preflight ---

total_missing=$((missing_count + missing_contract_count))

if [ "${total_missing}" -gt 0 ]; then
  echo "warmup: completed with missing dependencies (${total_missing}). see ${report_file}, ${agents_report_file}" >&2
  exit 2
fi

echo "warmup: completed. report: ${report_file}"
echo "warmup: agents report: ${agents_report_file}"
