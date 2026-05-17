# Mothership Skill Dependencies

This file defines external skills that should be available before active
orchestration proceeds beyond `intake`.

## Required Auxiliary Skills

- `obra/superpowers`
- `obra/brainstorming`
- `obra/making-plans`
- `obra/executing-plans`
- `obra/systematical-debugging`

These are preinstalled in this repository under:
- `skills/obra/superpowers/SKILL.md`
- `skills/obra/brainstorming/SKILL.md`
- `skills/obra/making-plans/SKILL.md`
- `skills/obra/executing-plans/SKILL.md`
- `skills/obra/systematical-debugging/SKILL.md`

## Phase Mapping

- `intake`: `task-intake-and-decomposition` and `obra/brainstorming`
- `research`: `obra/brainstorming` (research mode), `obra/superpowers`, `obra/making-plans`, and `research-execution`
- `planning`: `obra/brainstorming` (interactive mode) and `obra/making-plans`
- `coding`: `obra/executing-plans`; after QA requested changes, also use `obra/systematical-debugging`

## Commander Rule

During startup warm-up, commander should verify these skills are installed. If
any are missing, record `blocked` overlay details or continue with explicit
fallback notes when safe.
