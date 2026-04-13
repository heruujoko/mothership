# Mothership Skill Dependencies

This file defines external skills that should be available before active
orchestration proceeds beyond `intake`.

## Required Auxiliary Skills

- `obra/superpowers`
- `obra/creating-plan`
- `obra/systematical-debugging`

These are preinstalled in this repository under:
- `skills/obra/superpowers/SKILL.md`
- `skills/obra/creating-plan/SKILL.md`
- `skills/obra/systematical-debugging/SKILL.md`

## Phase Mapping

- `research`: prefer `obra/superpowers` and `obra/creating-plan`
- `coding` after QA requested changes: prefer `obra/systematical-debugging`

## Commander Rule

During startup warm-up, commander should verify these skills are installed. If
any are missing, record `blocked` overlay details or continue with explicit
fallback notes when safe.
