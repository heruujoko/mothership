# Mothership Wiki Design

Date: 2026-05-01
Status: Draft

## Problem

Mothership writes checkpoints per state transition but never reads them back.
Every task starts from scratch — no accumulated knowledge about the project,
no cross-referencing between tasks, no memory of past decisions or discoveries.

The Karpathy LLM Wiki pattern (https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)
provides a proven model: a persistent, compounding wiki maintained by the LLM
that sits between raw sources and schema configuration.

## Summary

Add a wiki layer to mothership that compounds knowledge across sessions. The
wiki uses a two-tier knowledge lifecycle (L1 draft, L2 verified), per-project
namespacing, and automatic ingestion at task completion.

## Knowledge Tiers

**L1 — Draft/Unverified (`.draft/`)**
- Written during any workflow state by any role
- Append-only, no citation required
- Stored in `.draft/<task-id>/` per active task
- `<task-id>` is derived from the checkpoint naming convention (e.g.,
  `2026-05-01-opencode-installer` from the intake checkpoint)
- Discarded if not promoted within 2 completed tasks

**L2 — Verified (insights/, decisions/)**
- Promoted from L1 only at `complete` by commander
- Requires citation (PR URL, commit SHA, or issue number)
- Merged into existing wiki pages (not new files per task)
- Permanent, cross-linked, indexed

## Directory Structure

```
<wiki-root>/                            # configured via /mothership:setup
  index.md                              # master catalog (L2 only)
  log.md                                # global append-only timeline
  schema.md                             # wiki conventions and rules
  projects/
    <project-name>/
      index.md                          # project-specific catalog
      log.md                            # project-specific timeline
      .draft/                           # L1 — draft/unverified
        <task-id>/
          insights.md                   # discoveries during task
          decisions.md                  # preliminary decisions
          notes.md                      # raw observations
      insights/                         # L2 — verified, promoted
        architecture.md
        patterns.md
        gotchas.md
        <topic>.md
      decisions/                        # L2 — verified, promoted
        <NNN>-<slug>.md
```

The `.draft/` directory is dot-prefixed so it is excluded from version control
and does not clutter wiki browsing. It is working memory, not permanent
knowledge.

## `/mothership:setup` Skill

A new skill that configures the mothership wiki for the current project.

**Responsibilities:**
1. Ask where the wiki root should live (default: `~/.mothership/wiki/`)
2. Register the current project name in the wiki
3. Write preferences to `.mothership/wiki.yaml` in the project root
4. Create the wiki directory structure if it does not exist
5. Initialize `schema.md` with default conventions

**Preferences stored in `.mothership/wiki.yaml`:**
```yaml
wiki:
  root: ~/.mothership/wiki/
  project: table-top
  auto_ingest: true
  draft_ttl: 2
```

Invoked via `/mothership:setup`. Runs once per project, re-runnable to change
preferences.

## Operations

### Ingest (automatic at `complete`)

Phase 1 — L1 Capture (during task lifecycle):
- Each workflow state appends to `.draft/<task-id>/insights.md` or
  `.draft/<task-id>/decisions.md`
- Research discovers something: append to L1 insights
- Planning makes a decision: append to L1 decisions
- Coder finds a gotcha: append to L1 insights
- Lightweight, no citation required

Phase 2 — L2 Promotion (at `complete`):
- Commander reads all L1 items from `.draft/<task-id>/`
- Quality gate criteria — an item promotes if ALL of:
  1. It has at least one verifiable citation (PR URL, commit SHA, or issue)
  2. It is not a duplicate of existing L2 content
  3. It survived QA (not challenged or superseded during review)
  4. It describes something useful to a future session (not task-specific trivia)
- Promoted items are merged into existing wiki pages
- Each promoted item must include citation
- `index.md` and `log.md` are updated
- `.draft/<task-id>/` is removed after promotion

### Query (at `intake` and `research`)

- At `intake`: commander reads `projects/<name>/index.md` for context
- At `research`: researcher reads relevant insight/decision pages
- No RAG, no vector search — structured markdown with index-based lookup
- Commander may read insights from other projects when relevant

### Lint (two tiers)

Lightweight (at `intake`):
- Check for stale `.draft/` folders (exceeded TTL)
- Check `index.md` references point to real files
- Flag contradictions between recent decisions

Full (`/mothership:lint`):
- All lightweight checks
- Scan for orphan pages (no inbound references)
- Cross-reference decisions against closed/merged PRs (citation validation)
- Detect duplicate insights across pages
- Report: contradictions, orphans, stale drafts, broken citations

## Workflow Integration

No new states. Wiki operations are side-effects on the existing state machine.

| State         | Wiki action                                        |
|---------------|----------------------------------------------------|
| `intake`      | Read `index.md` for context. Run lightweight lint. |
| `research`    | Append findings to `.draft/<task-id>/insights.md`  |
| `planning`    | Append decisions to `.draft/<task-id>/decisions.md`|
| `coding`      | Append discoveries to `.draft/<task-id>/insights.md`|
| `qa`          | Confirm or challenge L1 items                      |
| `complete`    | Run ingest: L1 to L2 promotion, update index/log   |

New exit criterion for `complete`:
- Wiki ingest is recorded (items promoted count + items discarded count)
- Index and log are updated
- `.draft/<task-id>/` is cleaned up

Role contracts get one addition each:
- Researcher: append discoveries to L1 staging
- Coder: append implementation discoveries to L1 staging
- QA: note which L1 items were confirmed or challenged
- Commander: run L2 promotion at complete

## Wiki Protocol (`skills/mothership/wiki-protocol.md`)

1. Commander owns L2 writes. Sub-agents write to L1 staging only.
2. L1 writes are append-only with headers: `### [state] <summary>`.
3. L2 promotion is a quality gate at `complete`.
4. Query is read-only for any role. Only commander writes L2.
5. Lint is commander-only, run at `intake` or on explicit invocation.
6. Cross-project queries allowed when context is relevant.

## Schema Conventions (`schema.md`)

Lives at `<wiki-root>/schema.md` — wiki-wide, read once at startup.

**Writing rules:**
- Insights pages use `## <topic>` sections within a single file
- Decisions use one file per decision with frontmatter: `id`, `date`,
  `status`, `citations`, `alternatives_considered`
- Every L2 claim must cite a source:
  `> Source: PR #5 (https://github.com/...)`
- Use `[[wikilinks]]` for cross-references between pages
- Append to existing pages, do not create new files for existing topics

**L1 to L2 promotion rules:**
- An item must have at least one citation to promote
- Contradictions with existing L2 content are flagged, not silently overwritten
- If a new insight challenges an existing one, create a `## Contradiction`
  section in both pages citing each other

**Index format:**
```markdown
# <Project> Index

## Insights
- [[architecture]] -- How the codebase is structured
- [[patterns]] -- Recurring patterns
- [[gotchas]] -- Non-obvious traps

## Decisions
- [[001-use-map-lookups]] -- Replace || chains with map lookups
- [[002-opencode-target]] -- Add opencode as installer target
```

**Log format:**
```markdown
## [2026-05-01] ingest | Add opencode installer target
Promoted 3 insights, 1 decision from task mship-opencode.
Citations: PR #5, commit fccf355
```

## Files to Create or Modify

New files:
- `skills/mothership/wiki-protocol.md` -- wiki operation contract
- `skills/mothership/scripts/wiki-setup.sh` -- setup script (or skill)
- `mothership-config/wiki-schema-default.md` -- default schema.md content

Modified files:
- `mothership-config/workflow.yaml` -- add wiki exit criteria to states
- `mothership-config/roles.md` -- add wiki responsibilities per role
- `agents/researcher.md` -- add L1 staging writes
- `agents/coder.md` -- add L1 staging writes
- `agents/qa.md` -- add L1 confirmation/challenge notes
- `agents/commander.md` -- add L2 promotion at complete
- `skills/mothership/SKILL.md` -- reference wiki operations
