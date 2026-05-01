# Wiki Schema

This file defines conventions for reading and writing the mothership wiki.
The LLM reads this file at startup to understand wiki structure and rules.

## Writing Rules

- Insights pages use `## <topic>` sections within a single file
- Decisions use one file per decision with YAML frontmatter:
  `id`, `date`, `status`, `citations`, `alternatives_considered`
- Every L2 claim must cite a source:
  `> Source: PR #5 (https://github.com/...)`
- Use `[[wikilinks]]` for cross-references between pages
- Append to existing pages, do not create new files for existing topics

## Decision File Template

```
---
id: NNN
date: YYYY-MM-DD
status: accepted
citations:
  - PR #N (url)
alternatives_considered:
  - <alternative 1>
  - <alternative 2>
---

## Context
<why this decision was needed>

## Decision
<what was decided>

## Consequences
<what changes as a result>
```

## L1 to L2 Promotion Rules

- An item must have at least one citation to promote
- Contradictions with existing L2 content are flagged, not silently overwritten
- If a new insight challenges an existing one, create a `## Contradiction`
  section in both pages citing each other

## Index Format

```markdown
# <Project> Index

## Insights
- [[architecture]] -- How the codebase is structured
- [[patterns]] -- Recurring patterns
- [[gotchas]] -- Non-obvious traps

## Decisions
- [[001-slug]] -- One-line summary
```

## Log Format

```markdown
## [YYYY-MM-DD] ingest | <summary>
Promoted N insights, M decisions from task <task-id>.
Citations: PR #X, commit SHA
```
