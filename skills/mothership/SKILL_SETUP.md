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
