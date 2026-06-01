# Mothership Local Agent Instructions

This repository is an installable mothership package for Codex, Claude, and Pi agents and skills.

## Entrypoint

- Use the `mothership` skill as the manual entrypoint when orchestration behavior is desired.
- Do not assume mothership should activate implicitly in mixed-skill environments.

## Supported Hosts

- **Claude Code** — uses built-in `Agent` tool for delegation
- **Codex** — uses built-in `spawn_agent` for delegation
- **Pi** — uses `mothership_spawn` extension tool with team-first delegation (requires `skills/mothership/extensions/subagent.ts` and `skills/mothership/extensions/pi-routing.js` installed under `~/.agents/extensions/`).

The Pi extension supports two delegation paths:

1. **Team path (preferred)**: When Pi-crew is installed, uses the `team` tool with model routing via `mothership-config/model-policy.yaml`. This enables consistent runtime state, reduced subprocess overhead, and model tier routing based on role requirements.

2. **Legacy path (fallback)**: Falls back to subprocess spawning when `team` is unavailable, with explicit logging of the reason.

Install with `./install.sh --target <host>`. See `agents/registry.yaml` for per-host role mappings.

## Canonical Config

- Roles registry: `mothership-config/roles.md`
- Workflow state machine: `mothership-config/workflow.yaml`
- Model policy: `mothership-config/model-policy.yaml` (fallback: warn + legacy model inheritance)
- Hub contract: `.mothership/hub/README.md`
- Skill dependencies: `mothership-config/skill-dependencies.md`

## Core Role Docs

- `agents/commander.md`
- `agents/researcher.md`
- `agents/coder.md`
- `agents/qa.md`
- `agents/pr_monkey.md`

## Supporting Skills

- `skills/commit/SKILL.md`
- `skills/commit-push/SKILL.md`
- `skills/create-pr/SKILL.md`
- `skills/pr-maintainer/SKILL.md`
- `skills/task-intake-and-decomposition/SKILL.md`
- `skills/risk-assessment/SKILL.md`
- `skills/research-execution/SKILL.md`
- `skills/parallelization-decision/SKILL.md`
- `skills/github-issue-management/SKILL.md`
- `skills/worktree-management/SKILL.md`
- `skills/qa-review-loop/SKILL.md`
- `skills/human-escalation/SKILL.md`
- `skills/obra/superpowers/SKILL.md`
- `skills/obra/brainstorming/SKILL.md`
- `skills/obra/making-plans/SKILL.md`
- `skills/obra/executing-plans/SKILL.md`
- `skills/obra/systematical-debugging/SKILL.md`
- `skills/model-policy-setup/SKILL.md`

## Intake Warm-Up

- Run `skills/mothership/scripts/warmup.sh` during `intake`.
- Warm-up validates `model-policy.yaml` shape when present; missing/invalid policy warns and falls back to legacy model inheritance.
- Keep runtime artifacts under `.mothership/` and ignored by git.

## Configuration

After installation, configure model routing for Pi (required for team-first routing):

```bash
# Initialize model-policy.yaml (if not auto-created by warmup)
skills/model-policy-setup/scripts/setup-model-policy.sh

# Edit the configuration
vim mothership-config/model-policy.yaml
```

### Model Policy Configuration

The [model-policy.yaml](mothership-config/model-policy.yaml) file controls:

- **host aliases**: Map local host names to supported targets (`claude`, `codex`, `pi`, `hermes`)
- **role tiers**: Assign model tiers per role (`commander`, `researcher`, `coder`, `qa`, `pr_monkey`)
- **profile overrides**: Adjust tier selection based on execution profile (`quick`/`standard`/`strict`)
- **risk overrides**: Adjust tier selection based on task risk (`low`/`medium`/`high`)
- **host models**: Define available models per host and tier
- **fallback behavior**: How to handle missing/invalid config

#### Example Configuration

```yaml
# mothership-config/model-policy.yaml

# Host aliases - map local host names to supported targets
host_aliases:
  claude: claude
  codex: codex
  pi: pi
  hermes: hermes

# Role tiers - map each role to a model tier
role_tiers:
  commander: high
  researcher: medium
  coder: medium
  qa: low
  pr_monkey: low

# Profile overrides - adjust tier based on workflow ceremony
profile_overrides:
  quick:
    researcher: low
    coder: low
    qa: low
  standard: {}
  strict:
    commander: high
    researcher: high
    coder: high
    qa: high

# Risk overrides - adjust tier based on task risk
risk_overrides:
  low:
    coder: low
    qa: low
  medium: {}
  high:
    researcher: high
    coder: high
    qa: high

# Host models - available models per host and tier
host_models:
  pi:
    high: ["claude-3-5-sonnet-latest", "claude-3-5-haiku-latest"]
    medium: ["claude-3-5-sonnet-latest", "claude-3-5-haiku-latest"]
    low: ["claude-3-5-haiku-latest"]
  codex:
    high: ["claude-3-5-sonnet-latest"]
    medium: ["claude-3-5-haiku-latest"]
    low: ["claude-3-5-haiku-latest"]
```

Model resolution order:
1. Explicit invocation override, if provided
2. Profile override from `profile_overrides`
3. Risk override from `risk_overrides`
4. Role tier from `role_tiers`
5. Host-specific model mapping in `agents/registry.yaml`
6. Fallback: inherit from current session model

### Verifying Configuration

```bash
# Check model policy exists
if [ -f "mothership-config/model-policy.yaml" ]; then
  echo "Model policy found"
else
  echo "Model policy missing - run model-policy-setup skill or script"
fi

# Validate YAML syntax
python3 -c "import yaml; yaml.safe_load(open('mothership-config/model-policy.yaml'))" 2>/dev/null || echo "YAML syntax error"
```
