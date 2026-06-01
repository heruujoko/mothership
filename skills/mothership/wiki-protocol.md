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

## Retrieval Protocol

### Purpose
Ensure that at `intake`, the commander reads only relevant wiki pages instead of the entire wiki. This reduces token waste and avoids noise from unrelated pages.

### Protocol (Semantic Prefix Search)

At `intake`, before reading wiki pages:

1. **Extract keywords** from the task title and body: categories, project domains, technical terms.
2. **Search by category + tags first:** Scan `projects/<name>/index.md` for matching category, tag, or keyword.
3. **Search by content:** If prefix search yields < 3 results, scan insight page titles and metadata (`tags` field) for relevance.
4. **Limit:** Return at most 5 pages. Do not read the full wiki.

Relevance signals, in priority order:
1. **Category match** — page is in a category that matches the task domain
2. **Tag overlap** — page tags intersect with task keywords
3. **Title keyword match** — page title contains a task keyword

### Implementation in commander (`intake`)

```yaml
wiki_retrieval:
  method: semantic_prefix_search  # search by tags/category first, then content
  max_pages: 5
  relevance_signals: [category, tags, title_keywords]
```

The commander should:
1. Read the project `index.md` for page list.
2. Filter by task-relevant categories and tags.
3. Read only the matching pages (up to 5).
4. If no match found, skip wiki read — do not default to reading everything.

## Hygiene Automation

### Stale Detection (at `intake`)

When wiki is configured, the commander checks page freshness during `intake`:

1. Read `freshness` field from each page's metadata header.
2. Compare against current date. If older than `stale_threshold` (default: 90 days), flag as stale.
3. Record stale pages in hub state under `wiki.stale_pages`.
4. Do not block work on stale pages — flag them for later cleanup.

**Configuration** (add to `.mothership/wiki.yaml`):

```yaml
wiki:
  stale_threshold_days: 90
  archive_after_days: 120  # pages older than this can be auto-archived
```

### Duplicate Detection (at promotion)

Before promoting an L1 item to L2, the commander:

1. Extracts key headings and content from the L1 draft.
2. Compares against existing L2 page headings and first paragraphs.
3. If heading overlap > 70% OR content similarity > 60%, flag as potential duplicate.
4. Do NOT create the duplicate — merge the new content into the existing page instead.
5. Record the merge decision in the promotion log.

### Cleanup Command: `kb stale`

Add a command `mothership kb stale` that:

1. Scans all pages under `projects/<name>/insights/` and `projects/<name>/decisions/`.
2. Reads the `freshness` field from each page's metadata header.
3. Lists pages where freshness is older than `stale_threshold_days`.
4. Reports: page path, freshness date, days since last update.

### Archival (not deletion)

When a page exceeds `archive_after_days`:

1. Move the page from `insights/` or `decisions/` to `<wiki-root>/archive/`.
2. Preserve the original metadata and full content.
3. Update `index.md` to note the archive location.
4. Never delete content — only archive it.

## Lint Rules

### Lightweight (at `intake`)
- Check for stale `.draft/` folders (exceeded `draft_ttl`)
- Check `index.md` references point to real files
- Flag contradictions between recent decisions
- **Read up to 5 relevant wiki pages (retrieval protocol)**
- **Check freshness metadata and flag stale pages**

### Full (`/mothership:lint`)
- All lightweight checks
- Scan for orphan pages (no inbound references)
- Cross-reference decisions against closed/merged PRs
- Detect duplicate insights across pages
- **Scan for pages older than `archive_after_days` and recommend archival**
- Report: contradictions, orphans, stale drafts, broken citations, stale pages
