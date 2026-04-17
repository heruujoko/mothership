# Mothership Local Agent Instructions

This repository is an installable mothership package for Codex agents and skills.

## Entrypoint

- Use the `mothership` skill as the manual entrypoint when orchestration behavior is desired.
- Do not assume mothership should activate implicitly in mixed-skill environments.

## Canonical Config

- Roles registry: `mothership-config/roles.md`
- Workflow state machine: `mothership-config/workflow.yaml`
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
- `skills/task-intake-and-decomposition.md`
- `skills/risk-assessment.md`
- `skills/research-execution.md`
- `skills/parallelization-decision.md`
- `skills/github-issue-management.md`
- `skills/worktree-management.md`
- `skills/qa-review-loop.md`
- `skills/human-escalation.md`
- `skills/obra/superpowers/SKILL.md`
- `skills/obra/creating-plan/SKILL.md`
- `skills/obra/systematical-debugging/SKILL.md`

## Intake Warm-Up

- Run `skills/mothership/scripts/warmup.sh` during `intake`.
- Keep runtime artifacts under `.mothership/` and ignored by git.
