#!/usr/bin/env bash
set -euo pipefail

repo_root="${1:-$(pwd)}"
allow_missing="${ALLOW_MISSING_POLICY:-0}"
warn_only="${WARN_ONLY_POLICY:-1}"
policy_file="${repo_root}/mothership-config/model-policy.yaml"

warn() {
  echo "model-policy: WARN: $*" >&2
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

for host in claude codex pi hermes; do
  if ! rg -n "^  ${host}:" "${policy_file}" >/dev/null 2>&1; then
    warn "host alias missing: ${host}"
    valid=0
  fi
done

for role in commander researcher coder qa pr_monkey; do
  if ! rg -n "^  ${role}:" "${policy_file}" >/dev/null 2>&1; then
    warn "role_tiers missing role: ${role}"
    valid=0
  fi
done

for risk in low medium high; do
  if ! rg -n "^  ${risk}:" "${policy_file}" >/dev/null 2>&1; then
    warn "risk_overrides missing bucket: ${risk}"
    valid=0
  fi
done

# check host_models shape by extracting host_models block only
host_models_block="$({ awk '/^host_models:/{flag=1;next}/^[A-Za-z_]+:/{if(flag)exit}flag{print}' "${policy_file}"; } || true)"
if [ -z "${host_models_block}" ]; then
  warn "host_models block missing or empty"
  valid=0
else
  for host in claude codex pi hermes; do
    host_block="$({ awk -v h="${host}" '
      $0 ~ "^  "h":$" {flag=1; next}
      flag && $0 ~ "^  [a-zA-Z0-9_-]+:$" {exit}
      flag {print}
    ' <<< "${host_models_block}"; } || true)"

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
