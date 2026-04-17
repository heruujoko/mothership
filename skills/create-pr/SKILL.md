---
name: create-pr
description: Create a pull request for the current branch with a real title and Markdown description after the branch exists on the remote. Use when the user asks to open a PR, draft a PR, or create a pull request description from the current changes.
---

# Create PR

Use this skill when the branch is ready for pull request creation and the user
wants a proper title and body.

If the branch is not pushed yet, stop and route to `commit-push` or the full
publish flow first.

## Preferred Tools

- Prefer the GitHub connector or app when repository and branch context are
  already known and PR creation is supported there.
- Use `gh pr create` as the fallback for current-branch PR creation.

## Preconditions

- Confirm the current branch is pushed to a remote.
- Resolve the repository, head branch, and base branch.
- Inspect the diff against the intended base before writing the PR text.

## PR Body Standard

The description should use real Markdown and cover:

- what changed
- why it changed
- user or developer impact
- root cause when the PR is a fix
- validation performed

Do not invent testing or impact statements. If validation is missing, say so
plainly.

## Workflow

1. Resolve repo, head branch, and base branch.
2. Inspect the diff and recent commits for an accurate summary.
3. Draft the PR title.
   - Keep it concise and representative of the full diff.
   - Default to a draft PR unless the user explicitly asks for ready review.
4. Draft a Markdown body with the required sections.
5. Create the PR through the connector or `gh pr create`.
6. Return the PR URL, title, base/head branches, and validation notes.

## Safety Rules

- Never open a PR from unpushed local commits.
- Never mark the PR ready for review unless the user asks.
- Never claim checks passed unless they actually ran and passed.
- If repository or auth context is missing, stop and report the blocker.
