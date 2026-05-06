---
name: parallelization-decision
description: Help commander decide whether work should be assigned to one coder agent or multiple coder agents in parallel.
---

# Parallelization Decision

## Purpose

Help commander decide whether work should be assigned to one coder agent or multiple coder agents.

## Inputs

- task decomposition
- file ownership overlap risk
- expected merge conflict risk
- machine resource snapshot
- urgency versus coordination cost

## Decision Rule

Use multiple coder agents only when:
- sub-tasks are independently scorable
- likely file overlap is low
- machine resources are sufficient
- coordination cost is justified

Use a single coder agent when:
- the task is tightly coupled
- sub-issues would touch the same files
- machine resources are constrained
- the work is small enough that parallelism would be slower overall

## Output

- single-agent or multi-agent decision
- rationale
- worktree allocation recommendation
