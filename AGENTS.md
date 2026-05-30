# Mothership Local Agent Instructions

This repository is an installable mothership package for Codex, Claude, and Pi agents and skills.

## Entrypoint

- Use the `mothership` skill as the manual entrypoint when orchestration behavior is desired.
- Do not assume mothership should activate implicitly in mixed-skill environments.

## Supported Hosts

- **Claude Code** — uses built-in `Agent` tool for delegation
- **Codex** — uses built-in `spawn_agent` for delegation
- **Pi** — uses `mothership_spawn` extension tool (requires `skills/mothership/extensions/subagent.ts` and `skills/mothership/extensions/pi-routing.js` installed under `~/.agents/extensions/`). The extension prefers direct `team` tool delegation when configured/available, otherwise falls back to legacy subprocess spawning with explicit reason logging. Set `PI_HOME` env var to override the install root.

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
- `agents/pr-monkey.md`

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
