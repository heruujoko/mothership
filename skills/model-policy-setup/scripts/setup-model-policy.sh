#!/usr/bin/env bash
set -euo pipefail

force=0
if [ "${1:-}" = "--force" ]; then
  force=1
fi

script_dir="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "${script_dir}/../../.." && pwd)"
policy_file="${repo_root}/mothership-config/model-policy.yaml"

if [ -f "${policy_file}" ] && [ "${force}" -ne 1 ]; then
  echo "model-policy-setup: ${policy_file} exists; skipping (use --force to overwrite)"
  exit 0
fi

mkdir -p "$(dirname "${policy_file}")"
cat > "${policy_file}" <<'YAML'
# Issue #17: model policy defaults and fallback behavior
version: 1

fallback:
  on_missing: "warn_and_inherit_current_thread_model"
  on_invalid: "warn_and_inherit_current_thread_model"
  legacy_behavior: "When no valid policy is available, delegation keeps legacy model inheritance."
  resolution_order:
    - explicit_override
    - profile_override
    - risk_override
    - role_tier
    - host_tier_default
    - inherit_current_thread_model

host_aliases:
  claude: claude
  codex: codex
  pi: pi
  hermes: hermes
  ollama: ollama

# Default tier by role.
# Goal: high-model planning/research, medium/low implementation.
role_tiers:
  commander: high
  researcher: high
  coder: medium
  qa: medium
  pr_monkey: low


# Optional profile-based overrides. Applied after explicit invocation overrides
# and before risk/role defaults so execution profiles can right-size ceremony.
profile_overrides:
  quick:
    commander: medium
    researcher: low
    coder: low
    qa: low
    pr_monkey: low
  standard: {}
  strict:
    commander: high
    researcher: high
    coder: high
    qa: high
    pr_monkey: medium

# Optional risk-based overrides.
risk_overrides:
  low:
    coder: low
    qa: low
  medium: {}
  high:
    researcher: high
    coder: high
    qa: high

# Host-specific model mapping by tier. Empty string means host default.
host_models:
  claude:
    high: "claude-opus-4"
    medium: "claude-sonnet-4"
    low: "claude-haiku-3.5"
  codex:
    high: "codex/gpt-5.5"
    medium: "codex/gpt-5.3-codex"
    low: "codex/gpt-5-mini"
  pi:
    high: "openai-codex/gpt-5.5"
    medium: "openai-codex/gpt-5.3-codex"
    low: "openai-codex/gpt-5-mini"
  hermes:
    high: ""
    medium: ""
    low: ""
  ollama:
    high: "codex/gpt-5.5"
    medium: "ollama/deepseek-v4-flash"
    low: "ollama/kimi-k2.6"

tiers:
  high:
    description: "Highest reasoning tier for orchestration and complex implementation."
    default_model: high
  medium:
    description: "Balanced tier for structured execution and reviews."
    default_model: medium
  low:
    description: "Lightweight tier for routine or narrow tasks."
    default_model: low
YAML

echo "model-policy-setup: wrote ${policy_file}"
