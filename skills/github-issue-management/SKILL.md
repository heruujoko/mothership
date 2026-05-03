# Skill: GitHub Issue Management

## Purpose

Define how mothership uses GitHub issues as durable workflow artifacts.

## Tracking Strategy

Default to using a single parent GitHub issue as the main tracker for a task.

- Use the parent issue as the primary tracker for the entire task lifecycle.
- Use issue comments for temporary notes: research findings, planning decisions, QA results, completion summaries.
- Only create child implementation issues when work is genuinely split into separately owned sub-scopes (different branches, different owners, different timelines).
- Do NOT create child issues just to track individual workflow phases within the same task.

## Issue Types

- research issue
- implementation issue
- escalation issue or comment
- completion summary comment

## Responsibilities

- create issues with clear scope
- link issues to parent task
- record research findings
- record PR references
- record QA findings
- record escalation status
- close issues when complete

## Rule

If a local hub decision materially changes task state, a GitHub artifact should also reflect it.

## Closure Hygiene

Before closing an issue (via `Closes #...` in a PR or manually):

1. Verify the issue body checklist still matches reality.
2. If any checklist items are stale (checked but not done, or unchecked but actually done), update the issue body or add a clarifying comment.
3. Never leave an issue in a "closed but unchecked checklist" state.

## Markdown Formatting Rule

All GitHub comments and updates must use proper Markdown formatting with headings, bullets, and sections. Never use literal `\n` escape sequences — use actual line breaks and Markdown structure.

## Template Usage

| Template | Use during workflow state |
|----------|--------------------------|
| Research Update | research → planning |
| Planning Update | planning → coding |
| QA Findings | qa → complete |
| Completion Summary | coding → review (or final) |

## Update Templates

### Research Update

```
## Research Update
- **Status**: [in progress | complete | blocked]
- **Findings**:
  - [finding 1]
  - [finding 2]
- **Open questions**:
  - [question 1]
- **Recommendation**: [summary]
```

### Planning Update

```
## Planning Update
- **Execution shape**: [single-agent | multi-agent]
- **Scope**: [summary of what will change]
- **Risks**: [risk list]
- **Issues**: [issue references]
```

### QA Findings

```
## QA Findings
- **Disposition**: [approved | changes requested | escalated]
- **Verified**:
  - [criterion 1]: [pass/fail]
- **Issues found**: [list or none]
- **Required changes**: [list or none]
```

### Completion Summary

```
## Completion Summary
- **Status**: [complete | partial]
- **What changed**: [summary]
- **PR**: [link]
- **Issue updates**: [what was updated]
```
