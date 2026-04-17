---
name: commit
description: Create a local git commit for a scoped set of changes without pushing. Use when the user asks to commit current work, make a commit message, or checkpoint changes locally but does not ask to push or open a pull request.
---

# Commit

Use this skill only for a local commit boundary.

Do not push. Do not open a PR. If the user wants either of those, route to
`commit-push` or `create-pr`.

## Preconditions

- Confirm the repository state with `git status -sb`.
- Inspect the relevant diff before staging.
- If the worktree is mixed, stage only the intended files. Do not silently
  include unrelated user changes.

## Workflow

1. Inspect status and diff for the requested scope.
2. Decide the staging strategy.
   - Prefer explicit file paths when scope is narrow.
   - Use `git add -A` only when the full worktree is clearly in scope.
3. Write an intentional commit message.
   - Default to a terse imperative summary.
   - Add a body only when the user asked for more context or the change needs
     rationale.
4. Commit locally.
5. Report the branch, short SHA, and what was included.

## Safety Rules

- Never amend or force-rewrite history unless the user explicitly asks.
- Never commit generated noise or unrelated untracked files without confirming
  they belong in scope.
- If hooks or tests fail, report the exact blocker instead of pretending the
  commit succeeded.

## Output Expectations

Return:

- branch name
- commit SHA
- commit message
- whether anything remains unstaged or uncommitted
