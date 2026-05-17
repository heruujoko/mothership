---
name: executing-plans
description: Bundled dependency skill used by mothership coding phase to execute an approved implementation plan step-by-step without skipping back into ad hoc design.
---

# obra/executing-plans

Use this skill at the start of `coding`, after `planning` has produced an
approved implementation plan.

## Intended Usage

- read the approved implementation plan before touching code
- execute the work in the plan's order unless a blocker forces escalation
- report plan mismatches or missing prerequisites instead of silently redesigning
- keep implementation notes tied back to the plan so QA can verify intent

## Mothership Routing

- In `coding`, coder agents should invoke this skill before implementation.
- If QA returns the task with defects or ambiguity, use
  `obra/systematical-debugging` first, then resume this skill.
- Do NOT use this skill during `intake`, `research`, or `planning`.

## Note

In environments with an upstream obra package, this local marker may be
replaced or shadowed by the upstream installation.
