---
name: commit-push
description: Commit scoped local changes and push the current branch to its remote without opening a pull request. Use when the user explicitly asks to commit and push, publish a branch, or update an existing remote branch but does not ask to create a PR.
---

# Commit Push

Use this skill for the local-to-remote publish step when the job ends at a
pushed branch.

Do not open a PR here. If the user also wants a PR, route to `create-pr` after
the push succeeds.

## Preconditions

- Check `git status -sb` and inspect the scoped diff.
- Confirm the current branch and remote configuration.
- If the current branch is the default branch and the user did not explicitly
  ask to push there, stop and clarify before pushing.

## Workflow

1. Inspect status, diff, branch, and upstream state.
2. Stage only the intended files.
3. Create an intentional commit message.
4. Commit locally.
5. Push the current branch.
   - If no upstream exists, use `git push -u origin $(git branch --show-current)`.
   - Otherwise push to the configured upstream.
6. Report the branch, commit SHA, remote target, and push result.

## Safety Rules

- Never push unrelated work from a mixed tree.
- Never force-push unless the user explicitly requests it.
- Never create a pull request inside this skill.
- If the remote is missing or rejects the push, stop and report the exact
  blocker.

## Output Expectations

Return:

- branch name
- commit SHA
- pushed remote and upstream
- whether the branch is now published remotely
