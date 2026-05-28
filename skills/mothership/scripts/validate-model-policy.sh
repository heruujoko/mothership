#!/usr/bin/env bash
set -euo pipefail

repo_root="${1:-$(pwd)}"
allow_missing="${ALLOW_MISSING_POLICY:-0}"
warn_only="${WARN_ONLY_POLICY:-1}"
policy_file="${repo_root}/mothership-config/model-policy.yaml"

warn() {
  echo "model-policy: WARN: $*" >&2
}

section_block() {
  local section="$1"
  awk -v section="${section}" '
    $0 ~ "^"section":$" { flag=1; next }
    flag && $0 ~ "^[A-Za-z_][A-Za-z0-9_]*:" { exit }
    flag { print }
  ' "${policy_file}"
}

nested_block() {
  local parent_block="$1"
  local key="$2"
  awk -v key="${key}" '
    $0 ~ "^  "key":$" { flag=1; next }
    flag && $0 ~ "^  [a-zA-Z0-9_-]+:$" { exit }
    flag { print }
  ' <<< "${parent_block}"
}

if [ ! -f "${policy_file}" ]; then
  warn "missing ${policy_file}; using legacy model inheritance (current thread model)."
  if [ "${allow_missing}" = "1" ]; then
    exit 0
  fi
  exit 1
fi

valid=1

for key in version fallback host_aliases role_tiers risk_overrides host_models tiers; do
  if ! rg -n "^${key}:" "${policy_file}" >/dev/null 2>&1; then
    warn "missing top-level key: ${key}"
    valid=0
  fi
done

# host_aliases (section-scoped)
host_aliases_block="$(section_block host_aliases || true)"
if [ -z "${host_aliases_block}" ]; then
  warn "host_aliases block missing or empty"
  valid=0
else
  for host in claude codex pi hermes; do
    if ! rg -n "^  ${host}:" <<< "${host_aliases_block}" >/dev/null 2>&1; then
      warn "host alias missing: ${host}"
      valid=0
    fi
  done
fi

# role_tiers (section-scoped)
role_tiers_block="$(section_block role_tiers || true)"
if [ -z "${role_tiers_block}" ]; then
  warn "role_tiers block missing or empty"
  valid=0
else
  for role in commander researcher coder qa pr_monkey; do
    if ! rg -n "^  ${role}:" <<< "${role_tiers_block}" >/dev/null 2>&1; then
      warn "role_tiers missing role: ${role}"
      valid=0
    fi
  done
fi

# risk_overrides (section-scoped)
risk_overrides_block="$(section_block risk_overrides || true)"
if [ -z "${risk_overrides_block}" ]; then
  warn "risk_overrides block missing or empty"
  valid=0
else
  for risk in low medium high; do
    if ! rg -n "^  ${risk}:" <<< "${risk_overrides_block}" >/dev/null 2>&1; then
      warn "risk_overrides missing bucket: ${risk}"
      valid=0
    fi
  done
fi

# host_models nested shape checks
host_models_block="$(section_block host_models || true)"
if [ -z "${host_models_block}" ]; then
  warn "host_models block missing or empty"
  valid=0
else
  for host in claude codex pi hermes; do
    host_block="$(nested_block "${host_models_block}" "${host}" || true)"

    if [ -z "${host_block}" ]; then
      warn "host_models missing host: ${host}"
      valid=0
      continue
    fi

    for tier in high medium low; do
      if ! rg -n "^    ${tier}:" <<< "${host_block}" >/dev/null 2>&1; then
        warn "host_models.${host} missing tier: ${tier}"
        valid=0
      fi
    done
  done
fi

if [ "${valid}" = "1" ]; then
  echo "model-policy: OK (${policy_file})"
  exit 0
fi

warn "invalid shape; using fallback: inherit current thread model"
if [ "${warn_only}" = "1" ]; then
  exit 0
fi
exit 2
