---
name: model-policy-setup
description: Bootstrap mothership model policy config. Creates mothership-config/model-policy.yaml if missing, or overwrite with --force.
---

# Model Policy Setup

Initializes `mothership-config/model-policy.yaml` with defaults for:
- host aliases (`claude`, `codex`, `pi`, `hermes`, `ollama`)
- role tiers (`commander`, `researcher`, `coder`, `qa`, `pr_monkey`)
- risk overrides (`low` / `medium` / `high`)
- host model mapping per tier (`high`, `medium`, `low`)
- fallback behavior (warn + legacy inheritance)

## Usage

```bash
skills/model-policy-setup/scripts/setup-model-policy.sh
skills/model-policy-setup/scripts/setup-model-policy.sh --force
```

- Default mode is non-destructive: does nothing if config already exists.
- `--force` overwrites existing config.

## Customization

After bootstrap, edit `mothership-config/model-policy.yaml`:
- change `role_tiers` per role
- tune `risk_overrides` (e.g., force `coder: low` on low-risk tasks)
- set `host_models.<host>.<tier>` to your available models
- remap `host_aliases` for local host naming
