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

## Post-Merge Cleanup

After a PR is merged:

1. Return to the main branch.
2. Pull latest changes.
3. Remove the worktree if it was temporary.
4. Confirm repo state is clean.

For the full post-merge reset sequence including next-task preparation, see `agents/commander.md` section "Post-Merge Reset".
