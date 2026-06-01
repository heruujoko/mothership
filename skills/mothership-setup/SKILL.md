---
name: mothership-setup
description: Configure the mothership wiki for the current project. Run once per project to set up wiki directory structure, preferences, and default schema. Re-run to change wiki root location or other settings.
---

# Mothership Wiki Setup

## What It Does

1. Prompts for wiki root location (default: `~/.mothership/wiki/`)
2. Derives project name from repository directory
3. Creates wiki directory structure under `<wiki-root>/projects/<name>/`
4. Writes preferences to `.mothership/wiki.yaml`
5. Initializes `schema.md` from `mothership-config/wiki-schema-default.md` if not present
6. Creates `index.md` and `log.md` for the project if not present
7. Bootstraps `mothership-config/model-policy.yaml` if missing (non-destructive)
8. Informs user where to edit model configuration for their crews
9. Appends setup entry to project and global logs

## Implementation

1. Executes `../mothership/scripts/wiki-setup.sh` from the skill directory.
2. Executes `../model-policy-setup/scripts/setup-model-policy.sh` to bootstrap model policy if missing.
3. Prints guidance on editing `mothership-config/model-policy.yaml` for crew model selection.

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
