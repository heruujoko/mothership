---
name: pr-maintainer
description: Maintain a GitHub pull request through the GitHub CLI by creating PRs, reading review feedback, routing each feedback item back through mothership, and updating the PR until it is ready for human-approved merge.
---

# PR Maintainer

Use this skill when pull request work itself becomes an ongoing maintenance
loop.

This skill is for:

- creating a PR from the current branch with a full description
- reading PR feedback and classifying it
- returning one feedback item at a time to commander as new scoped work
- replying to comments and resolving threads after the routed fix is complete
- repeating until the PR is ready for merge

This skill must operate through the `PR monkey` role contract in
`agents/pr-monkey.md`.

## Tooling Rule

Use `gh` for PR operations.

Before any PR work:

1. Check that `gh` is installed.
2. Check that `gh auth status` succeeds for the target repository.
3. If either check fails, stop and produce exact setup instructions for
   commander to relay to the user.
4. Continue only after `gh` is usable by the agent.

## Commander Integration

When mothership receives:

- a request to create a PR from the current project
- a PR URL or PR number for maintenance
- a request to read, triage, or respond to PR feedback

commander should spawn the `PR monkey` role instead of handling the PR details
itself.

PR monkey does not replace the main workflow. It extends it.

### External feedback handling

When PR monkey finds actionable feedback:

1. Read the feedback through `gh`.
2. Select one actionable item.
3. Summarize that item for commander.
4. Commander treats it as a new request and routes it through research,
   planning, coding, and QA as needed.
5. After the fix is complete, PR monkey pushes the follow-up branch state,
   replies to the comment, and resolves the thread if appropriate.
6. Then move to the next feedback item.

Do not batch multiple independent comments into one ambiguous task unless
commander explicitly re-scopes them together.

## Allowed Work

- create PRs with full Markdown bodies
- inspect reviews, comments, and thread state
- map comments to files, commits, and scoped follow-up tasks
- push already-approved fixes from the current branch
- reply and resolve review threads once the routed fix is present
- report remaining open feedback and merge readiness

## Prohibited Work

- do not solve review feedback directly outside commander workflow
- do not merge a PR without explicit human consent
- do not invent validation claims
- do not skip `gh` installation or auth checks
- do not hide unresolved feedback

## PR Body Standard

PR descriptions should include:

- what changed
- why it changed
- impact on users or developers
- root cause when the PR is a fix
- validation performed

## Output Expectations

Return concise, durable artifacts:

- PR URL and title when creating a PR
- one feedback item at a time when triaging review comments
- push result and thread reply status after a routed fix lands
- explicit list of remaining blockers, open feedback, and merge dependencies
