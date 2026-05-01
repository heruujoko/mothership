# Wiki Protocol

This document defines how wiki operations are performed across mothership roles.
Use it together with `mothership-config/workflow.yaml`, `skills/mothership/subagent-protocol.md`,
and the wiki design spec at `docs/superpowers/specs/2026-05-01-mothership-wiki-design.md`.

## Core Rule

Commander owns all L2 (verified) wiki writes. Sub-agents write to L1 staging only.
Query is read-only for any role.

## Knowledge Tiers

### L1 — Draft/Unverified (`.draft/`)

- Stored in `<wiki-root>/projects/<name>/.draft/<task-id>/`
- Written during any workflow state by any role
- Append-only with headers: `### [state] <summary>`
- No citation required
- Discarded if not promoted within `draft_ttl` completed tasks (default: 2)

### L2 — Verified (`insights/`, `decisions/`)

- Stored in `<wiki-root>/projects/<name>/insights/` and `decisions/`
- Promoted from L1 only at `complete` by commander
- Requires citation (PR URL, commit SHA, or issue number)
- Merged into existing wiki pages — not new files per task
- Permanent, cross-linked, indexed in `index.md`

## Role Permissions

| Role    | L1 Write | L2 Write | L2 Read | Lint |
|---------|----------|----------|---------|------|
| Commander | Yes    | Yes      | Yes     | Yes  |
| Researcher | Yes   | No       | Yes     | No   |
| Coder   | Yes      | No       | Yes     | No   |
| QA      | Yes      | No       | Yes     | No   |
| PR Monkey | Yes    | No       | Yes     | No   |

## Quality Gate Criteria

An L1 item promotes to L2 only when ALL of:
1. Has at least one verifiable citation (PR URL, commit SHA, or issue)
2. Is not a duplicate of existing L2 content
3. Survived QA (not challenged or superseded during review)
4. Describes something useful to a future session (not task-specific trivia)

## L1 Write Format

Each L1 file uses this format:

```
### [research] Discovered sidebar state bug
The WorkspaceViewModel does not clear sidebar state on connection switch.
Affected files: WorkspaceViewModel.kt

### [coding] Found that filepath.Dir uses OS-specific separators
On Windows, nested command paths break because backslash is used.

### [qa] Confirmed: sidebar state reset works across all connection types
Verified with regression tests in WorkspaceViewModelTest.
```

## L2 Promotion Process

At `complete`, commander:
1. Reads all files in `.draft/<task-id>/`
2. Applies quality gate criteria to each item
3. Merges promoted items into existing wiki pages:
   - Insights → append to the relevant topic page or create new section
   - Decisions → create numbered decision file
4. Updates `projects/<name>/index.md` with new entries
5. Appends to `projects/<name>/log.md` with promotion summary
6. Removes `.draft/<task-id>/` directory

## Cross-Project Queries

Commander may read insights from other projects when relevant.
Use `projects/<other-project>/index.md` to locate relevant pages.

## Lint Rules

### Lightweight (at `intake`)
- Check for stale `.draft/` folders (exceeded `draft_ttl`)
- Check `index.md` references point to real files
- Flag contradictions between recent decisions

### Full (`/mothership:lint`)
- All lightweight checks
- Scan for orphan pages (no inbound references)
- Cross-reference decisions against closed/merged PRs
- Detect duplicate insights across pages
- Report: contradictions, orphans, stale drafts, broken citations
