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
- selected execution profile (`quick`, `standard`, or `strict`) and rationale
- initial risk level
- research-needed flag
- possible implementation issue breakdown

## Execution Profile Classification

During intake, classify the task into exactly one profile and record the chosen profile plus rationale in hub state.

| Profile | Use when | Intake output requirements |
| --- | --- | --- |
| `quick` | Tiny fix, doc/content-only edit, cosmetic update, or single-line config with low blast radius | Record why minimal research and concise planning are sufficient; still require light QA. |
| `standard` | Normal feature, bugfix, or routine multi-file change | Record normal research/planning expectations and QA scope. |
| `strict` | Security, architecture, persistence, auth, crypto, secrets, hub/config, installer, or model-routing change | Record sensitive boundaries, deep research needs, high-model planning, and full QA/security review obligations. |

Auto-detection hints:

- One file or purely cosmetic/content changes usually indicate `quick`.
- Three or more files, behavior changes, or unclear coupling usually indicate `standard`.
- Keywords such as password, token, auth, crypto, secret, permission, or policy indicate `strict`.
- Touching `.mothership/hub/`, `mothership-config/`, persistence, serialization, install paths, or model routing indicates `strict`.
- If hints disagree, choose the stricter profile.

The profile controls ceremony, not state existence: every task still reaches QA, and `quick` only lightens research/planning when the skip/minimal rationale is recorded.

## Boundary Definition

During intake and decomposition, explicitly define the boundary between what the task changes and what it does not.

- If an interface implies data round-tripping (e.g., save/load, serialize/deserialize, persist/restore), note this as a boundary that QA must verify end-to-end.
- Record these boundaries in the task scope so both coder and QA agents can reference them.

## Key Rule

Do not split a task into parallel sub-issues when the split would cause shared ownership confusion or frequent merge conflicts.

## Cross-Reference

See `skills/github-issue-management/SKILL.md` for parent-issue-first tracking strategy.
