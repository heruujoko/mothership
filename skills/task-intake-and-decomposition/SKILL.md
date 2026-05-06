---
name: task-intake-and-decomposition
description: Help commander turn an incoming human request into an actionable execution plan. Covers scoping, decomposition, and success criteria definition.
---

# Task Intake and Decomposition

## Purpose

Help commander turn an incoming human request into an actionable execution plan.

## Responsibilities

- summarize the request
- identify success criteria
- identify missing information
- estimate complexity and coupling
- decide whether research is needed
- decide whether the task can be split safely

## Outputs

- task summary
- decomposition recommendation
- initial risk level
- research-needed flag
- possible implementation issue breakdown

## Boundary Definition

During intake and decomposition, explicitly define the boundary between what the task changes and what it does not.

- If an interface implies data round-tripping (e.g., save/load, serialize/deserialize, persist/restore), note this as a boundary that QA must verify end-to-end.
- Record these boundaries in the task scope so both coder and QA agents can reference them.

## Key Rule

Do not split a task into parallel sub-issues when the split would cause shared ownership confusion or frequent merge conflicts.

## Cross-Reference

See `skills/github-issue-management/SKILL.md` for parent-issue-first tracking strategy.
