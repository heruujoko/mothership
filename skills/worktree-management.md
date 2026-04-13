# Skill: Worktree Management

## Purpose

Define when and how isolated worktrees are used for coder agents.

## Use Worktrees When

- multiple coder agents are active
- branch isolation materially reduces interference
- disk and memory overhead are acceptable

## Avoid Worktrees When

- the task is small
- a single coder agent is sufficient
- machine resources are constrained

## Cleanup Rules

Before deleting a worktree:
- confirm branch state
- confirm uncommitted changes status
- confirm issue and PR state are recorded

Never silently remove a dirty worktree.
