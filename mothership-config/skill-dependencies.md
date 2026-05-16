# Mothership Skill Dependencies

This file defines external skills that should be available before active
orchestration proceeds beyond `intake`.

## Required Auxiliary Skills

- `obra/superpowers`
- `obra/creating-plan`
- `obra/systematical-debugging`
- `obra/brainstorming`

These are preinstalled in this repository under:
- `skills/obra/superpowers/SKILL.md`
- `skills/obra/creating-plan/SKILL.md`
- `skills/obra/systematical-debugging/SKILL.md`
- `skills/obra/brainstorming/SKILL.md`

## Phase Mapping

- `intake`: `task-intake-and-decomposition` for scoping
- `research`: `obra/superpowers` (research-mode brainstorming) and `research-execution`
- `planning`: `obra/brainstorming` (interactive-mode, if design exploration needed) and `writing-plans`
- `coding` after QA requested changes: `obra/systematical-debugging`

## Commander Rule

During startup warm-up, commander should verify these skills are installed. If
any are missing, record `blocked` overlay details or continue with explicit
fallback notes when safe.
