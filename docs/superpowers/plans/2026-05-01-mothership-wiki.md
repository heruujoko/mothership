# Mothership Wiki Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a persistent, compounding wiki layer to mothership with two-tier knowledge lifecycle (L1 draft, L2 verified), per-project namespacing, and automatic ingestion at task completion.

**Architecture:** The wiki is a new side-channel that compounds knowledge across sessions. It does not add new workflow states — instead it hooks into existing state transitions. Sub-agents write L1 drafts during research/coding/QA, and the commander promotes verified items to L2 at task completion. A `/mothership:setup` skill configures wiki location per project.

**Tech Stack:** Bash (setup script), Markdown (wiki content), YAML (preferences)

**Spec:** `docs/superpowers/specs/2026-05-01-mothership-wiki-design.md`

---

## File Map

| File | Responsibility |
|------|---------------|
| `skills/mothership/wiki-protocol.md` | Wiki operation contract — who writes what, when, quality gates |
| `skills/mothership/scripts/wiki-setup.sh` | Setup script for configuring wiki per project |
| `mothership-config/wiki-schema-default.md` | Default schema.md content for new wiki roots |
| `mothership-config/workflow.yaml` | Add wiki exit criteria to states |
| `mothership-config/roles.md` | Add wiki responsibilities section |
| `skills/mothership/SKILL.md` | Reference wiki operations in warm-up checklist and after-workflow steps |
| `skills/mothership/subagent-protocol.md` | Add wiki output shape for each role |
| `agents/researcher.md` | Add L1 staging writes |
| `agents/coder.md` | Add L1 staging writes |
| `agents/qa.md` | Add L1 confirmation/challenge notes |
| `agents/commander.md` | Add L2 promotion at complete, wiki query at intake |

---

### Task 1: Create wiki protocol document

**Files:**
- Create: `skills/mothership/wiki-protocol.md`

- [ ] **Step 1: Create the wiki protocol file**

```markdown
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
```

- [ ] **Step 2: Verify file exists and is readable**

Run: `head -5 skills/mothership/wiki-protocol.md`
Expected: Shows the first 5 lines of the protocol file.

- [ ] **Step 3: Commit**

```bash
git add skills/mothership/wiki-protocol.md
git commit -m "feat: add wiki protocol document for mothership knowledge layer"
```

---

### Task 2: Create wiki setup script

**Files:**
- Create: `skills/mothership/scripts/wiki-setup.sh`

- [ ] **Step 1: Create the setup script**

Model after `warmup.sh`. The script:
- Prompts for wiki root (default: `~/.mothership/wiki/`)
- Derives project name from repo directory name
- Writes `.mothership/wiki.yaml` preferences
- Creates wiki directory structure
- Initializes `schema.md` from default template

```bash
#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
project_name="$(basename "${repo_root})"
wiki_root_default="${HOME}/.mothership/wiki"
pref_file="${repo_root}/.mothership/wiki.yaml"

echo "Mothership Wiki Setup"
echo "====================="
echo "Project: ${project_name}"
echo

read -rp "Wiki root [${wiki_root_default}]: " wiki_root
wiki_root="${wiki_root:-${wiki_root_default}}"
wiki_root="$(eval echo "${wiki_root}")"

project_dir="${wiki_root}/projects/${project_name}"

mkdir -p "${project_dir}/.draft" "${project_dir}/insights" "${project_dir}/decisions"

mkdir -p "$(dirname "${pref_file}")"
cat > "${pref_file}" <<EOF
wiki:
  root: ${wiki_root}
  project: ${project_name}
  auto_ingest: true
  draft_ttl: 2
EOF

schema_file="${wiki_root}/schema.md"
if [ ! -f "${schema_file}" ]; then
  schema_template="${repo_root}/mothership-config/wiki-schema-default.md"
  if [ -f "${schema_template}" ]; then
    cp "${schema_template}" "${schema_file}"
  else
    cat > "${schema_file}" <<'SCHEMA'
# Wiki Schema

This file defines conventions for reading and writing the mothership wiki.

## Writing Rules

- Insights pages use `## <topic>` sections within a single file
- Decisions use one file per decision with frontmatter
- Every L2 claim must cite a source: `> Source: PR #N (url)`
- Use `[[wikilinks]]` for cross-references between pages
- Append to existing pages, do not create new files for existing topics

## L1 to L2 Promotion Rules

- An item must have at least one citation to promote
- Contradictions with existing L2 content are flagged, not silently overwritten
- If a new insight challenges an existing one, create a `## Contradiction` section in both pages

## Index Format

Each project `index.md` lists all insight pages and decision files with one-line summaries.

## Log Format

Entries use: `## [YYYY-MM-DD] ingest | <summary>` followed by promotion counts and citations.
SCHEMA
  fi
fi

global_index="${wiki_root}/index.md"
if [ ! -f "${global_index}" ]; then
  cat > "${global_index}" <<EOF
# Mothership Wiki Index

## Projects
- [[${project_name}]] — ${project_name} project wiki
EOF
fi

global_log="${wiki_root}/log.md"
if [ ! -f "${global_log}" ]; then
  echo "# Mothership Wiki Log" > "${global_log}"
fi

project_index="${project_dir}/index.md"
if [ ! -f "${project_index}" ]; then
  cat > "${project_index}" <<EOF
# ${project_name} Index

## Insights
_(no insights yet)_

## Decisions
_(no decisions yet)_
EOF
fi

project_log="${project_dir}/log.md"
if [ ! -f "${project_log}" ]; then
  echo "# ${project_name} Log" > "${project_log}"
fi

timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
cat >> "${project_log}" <<EOF

## [${timestamp}] setup | Wiki initialized
Root: ${wiki_root}
Project: ${project_name}
EOF

cat >> "${global_log}" <<EOF

## [${timestamp}] setup | ${project_name} wiki created
Root: ${wiki_root}
EOF

echo
echo "Wiki setup complete:"
echo "  root: ${wiki_root}"
echo "  project: ${project_name}"
echo "  project dir: ${project_dir}"
echo "  preferences: ${pref_file}"
echo "  schema: ${schema_file}"
```

- [ ] **Step 2: Make script executable**

Run: `chmod +x skills/mothership/scripts/wiki-setup.sh`

- [ ] **Step 3: Verify script syntax**

Run: `bash -n skills/mothership/scripts/wiki-setup.sh`
Expected: No output (clean parse).

- [ ] **Step 4: Commit**

```bash
git add skills/mothership/scripts/wiki-setup.sh
git commit -m "feat: add wiki setup script for mothership project configuration"
```

---

### Task 3: Create default wiki schema template

**Files:**
- Create: `mothership-config/wiki-schema-default.md`

- [ ] **Step 1: Create the schema template**

This is the template copied to `<wiki-root>/schema.md` on first setup.

```markdown
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
```

- [ ] **Step 2: Commit**

```bash
git add mothership-config/wiki-schema-default.md
git commit -m "feat: add default wiki schema template for mothership"
```

---

### Task 4: Update workflow.yaml with wiki exit criteria

**Files:**
- Modify: `mothership-config/workflow.yaml`

- [ ] **Step 1: Add wiki exit criteria to `intake` state**

In `mothership-config/workflow.yaml`, locate the `intake` exit_criteria list and add after the last item:

```yaml
      - "Wiki index is read for project context if wiki is configured."
```

- [ ] **Step 2: Add wiki exit criteria to `research` state**

In the `research` exit_criteria list, add:

```yaml
      - "Research findings are appended to `.draft/<task-id>/insights.md` if wiki is configured."
```

- [ ] **Step 3: Add wiki exit criteria to `planning` state**

In the `planning` exit_criteria list, add:

```yaml
      - "Planning decisions are appended to `.draft/<task-id>/decisions.md` if wiki is configured."
```

- [ ] **Step 4: Add wiki exit criteria to `coding` state**

In the `coding` exit_criteria list, add:

```yaml
      - "Implementation discoveries are appended to `.draft/<task-id>/insights.md` if wiki is configured."
```

- [ ] **Step 5: Add wiki exit criteria to `qa` state**

In the `qa` exit_criteria list, add:

```yaml
      - "L1 items confirmed or challenged by QA are noted in `.draft/<task-id>/notes.md` if wiki is configured."
```

- [ ] **Step 6: Add wiki exit criteria to `complete` state**

In the `complete` exit_criteria list, add:

```yaml
      - "Wiki ingest is recorded (items promoted count + items discarded count) if wiki is configured."
      - "Wiki index and log are updated with promoted items if wiki is configured."
      - "`.draft/<task-id>/` is cleaned up after promotion if wiki is configured."
```

- [ ] **Step 7: Verify YAML is valid**

Run: `python3 -c "import yaml; yaml.safe_load(open('mothership-config/workflow.yaml'))"`
Expected: No output (valid parse).

- [ ] **Step 8: Commit**

```bash
git add mothership-config/workflow.yaml
git commit -m "feat: add wiki exit criteria to workflow states"
```

---

### Task 5: Update roles.md with wiki responsibilities

**Files:**
- Modify: `mothership-config/roles.md`

- [ ] **Step 1: Add wiki responsibilities section**

Append a new section after the "Role Introduction Process" section:

```markdown
## Wiki Responsibilities

When wiki is configured for a project (via `/mothership:setup`), each role has additional wiki-related duties:

### Commander
- Reads `projects/<name>/index.md` at `intake` for project context
- Runs lightweight lint at `intake`
- Performs L2 promotion at `complete`: reads L1 drafts, applies quality gate, merges into wiki pages
- Updates `index.md` and `log.md` after promotion
- Cleans `.draft/<task-id>/` after promotion

### Research Agent
- Appends research discoveries to `.draft/<task-id>/insights.md` (L1)
- Format: `### [research] <summary>` followed by findings

### Coder Agent
- Appends implementation discoveries to `.draft/<task-id>/insights.md` (L1)
- Format: `### [coding] <summary>` followed by findings

### QA Agent
- Notes which L1 items were confirmed or challenged in `.draft/<task-id>/notes.md`
- Format: `### [qa] confirmed: <item>` or `### [qa] challenged: <item>`

### PR Monkey
- No specific wiki duties. PR context is captured by commander during L2 promotion.
```

- [ ] **Step 2: Commit**

```bash
git add mothership-config/roles.md
git commit -m "feat: add wiki responsibilities to roles registry"
```

---

### Task 6: Update subagent-protocol.md with wiki output shapes

**Files:**
- Modify: `skills/mothership/subagent-protocol.md`

- [ ] **Step 1: Add wiki output shapes**

After the existing "## Required Output Shapes" section, add L1 staging instructions to each role's output:

Add to the Researcher output section:

```markdown
### Wiki L1 Staging

When wiki is configured, the researcher should also append discoveries to
`.draft/<task-id>/insights.md` using the format:
`### [research] <summary>` followed by the finding.
This is in addition to (not instead of) the structured research output above.
```

Add to the Coder output section:

```markdown
### Wiki L1 Staging

When wiki is configured, the coder should also append implementation discoveries
to `.draft/<task-id>/insights.md` using the format:
`### [coding] <summary>` followed by the finding.
This includes non-obvious gotchas, patterns discovered, and architectural observations.
```

Add to the QA output section:

```markdown
### Wiki L1 Staging

When wiki is configured, QA should note which L1 items from `.draft/<task-id>/`
were confirmed or challenged in `.draft/<task-id>/notes.md`:
- `### [qa] confirmed: <item summary>` for items that survived review
- `### [qa] challenged: <item summary>` for items that are incorrect or superseded
```

Add to the PR Monkey output section — no changes needed since PR Monkey has no wiki duties.

- [ ] **Step 2: Commit**

```bash
git add skills/mothership/subagent-protocol.md
git commit -m "feat: add wiki L1 staging output shapes to sub-agent protocol"
```

---

### Task 7: Update agent contracts with wiki duties

**Files:**
- Modify: `agents/researcher.md`
- Modify: `agents/coder.md`
- Modify: `agents/qa.md`
- Modify: `agents/commander.md`

- [ ] **Step 1: Update researcher.md**

Add to the "Outputs" section (after the existing recommended output structure):

```markdown
- L1 wiki staging entries in `.draft/<task-id>/insights.md` when wiki is configured
```

Add to the "Boundaries" section:

```markdown
- The research agent does not write L2 wiki content. L1 staging only.
```

- [ ] **Step 2: Update coder.md**

Add to the "Outputs" section (after "blocker report when necessary"):

```markdown
- L1 wiki staging entries in `.draft/<task-id>/insights.md` when wiki is configured
```

Add to the "Completion Criteria" section (after the existing criteria):

```markdown
- implementation discoveries are appended to L1 staging if wiki is configured
```

- [ ] **Step 3: Update qa.md**

Add to the "Outputs" section (after "updated PR review comments with findings"):

```markdown
- L1 wiki confirmation/challenge notes in `.draft/<task-id>/notes.md` when wiki is configured
```

- [ ] **Step 4: Update commander.md**

Add to the "Responsibilities" list (after "Close issues and ensure cleanup is complete"):

```markdown
- Run wiki L2 promotion at `complete` when wiki is configured.
- Read wiki `index.md` for project context at `intake`.
```

Add to the "Outputs" list (after "Cleanup confirmation"):

```markdown
- Wiki ingest record (items promoted + discarded)
```

Add a new section after "Prohibited Behavior":

```markdown
## Wiki Operations

### At Intake
- Read `projects/<name>/index.md` for project context if wiki is configured.
- Run lightweight lint: check for stale drafts, broken index references.

### At Complete
- Read all files in `.draft/<task-id>/`.
- Apply quality gate criteria from `skills/mothership/wiki-protocol.md`.
- Merge promoted items into existing wiki pages.
- Update `projects/<name>/index.md` and `projects/<name>/log.md`.
- Remove `.draft/<task-id>/`.
- Record promotion counts in completion checkpoint.

### Wiki is Optional
When no wiki is configured (no `.mothership/wiki.yaml`), skip all wiki operations.
Do not prompt the user to set up wiki during an active task.
```

- [ ] **Step 5: Commit**

```bash
git add agents/researcher.md agents/coder.md agents/qa.md agents/commander.md
git commit -m "feat: add wiki L1 staging duties to agent contracts"
```

---

### Task 8: Update SKILL.md to reference wiki operations

**Files:**
- Modify: `skills/mothership/SKILL.md`

- [ ] **Step 1: Add wiki reference to warm-up checklist**

In the Startup Warm-Up Checklist, add after item 6 (the preflight failure check):

```markdown
7. If `.mothership/wiki.yaml` exists, read `projects/<name>/index.md` from the configured wiki root for project context.
```

- [ ] **Step 2: Add wiki reference to the purpose section**

In the "What Happens After Warmup" section, add after item 3 (follow state machine):

```markdown
4. If wiki is configured, L1 staging writes happen automatically during research, planning, coding, and QA states.
5. L2 promotion happens automatically at `complete`.
```

- [ ] **Step 3: Commit**

```bash
git add skills/mothership/SKILL.md
git commit -m "feat: reference wiki operations in mothership SKILL.md"
```

---

### Task 9: Add /mothership:setup skill document

**Files:**
- Create: `skills/mothership/SKILL_SETUP.md`

- [ ] **Step 1: Create the setup skill document**

This is the skill definition for `/mothership:setup`. It is a separate document referenced from SKILL.md.

```markdown
# Mothership Wiki Setup

## Purpose

Configure the mothership wiki for the current project. Sets up wiki directory
structure, preferences, and default schema.

## When to Use

Run `/mothership:setup` once per project to configure wiki preferences.
Re-run to change wiki root location or other settings.

## What It Does

1. Prompts for wiki root location (default: `~/.mothership/wiki/`)
2. Derives project name from repository directory
3. Creates wiki directory structure under `<wiki-root>/projects/<name>/`
4. Writes preferences to `.mothership/wiki.yaml`
5. Initializes `schema.md` from `mothership-config/wiki-schema-default.md` if not present
6. Creates `index.md` and `log.md` for the project if not present
7. Appends setup entry to project and global logs

## Implementation

Executes `skills/mothership/scripts/wiki-setup.sh` from the repository root.

## Preferences

Stored in `.mothership/wiki.yaml`:

```yaml
wiki:
  root: ~/.mothership/wiki/
  project: project-name
  auto_ingest: true
  draft_ttl: 2
```

- `root`: Absolute path to the wiki root directory
- `project`: Project name (used as directory name under `projects/`)
- `auto_ingest`: If true, L2 promotion happens automatically at `complete`
- `draft_ttl`: Number of completed tasks before stale L1 drafts are discarded
```

- [ ] **Step 2: Commit**

```bash
git add skills/mothership/SKILL_SETUP.md
git commit -m "feat: add /mothership:setup skill document"
```

---

### Task 10: Final verification

- [ ] **Step 1: Verify all new files exist**

Run: `ls -la skills/mothership/wiki-protocol.md skills/mothership/scripts/wiki-setup.sh mothership-config/wiki-schema-default.md skills/mothership/SKILL_SETUP.md`
Expected: All four files listed.

- [ ] **Step 2: Verify all modified files contain wiki references**

Run: `grep -l "wiki" mothership-config/workflow.yaml mothership-config/roles.md skills/mothership/subagent-protocol.md agents/researcher.md agents/coder.md agents/qa.md agents/commander.md skills/mothership/SKILL.md`
Expected: All seven files listed.

- [ ] **Step 3: Verify workflow YAML is valid**

Run: `python3 -c "import yaml; yaml.safe_load(open('mothership-config/workflow.yaml'))"`
Expected: No output (valid parse).

- [ ] **Step 4: Verify setup script is executable and parses cleanly**

Run: `bash -n skills/mothership/scripts/wiki-setup.sh && echo "OK"`
Expected: `OK`

- [ ] **Step 5: Verify git log shows all commits**

Run: `git log --oneline main..HEAD`
Expected: 9 commits (tasks 1-9).
