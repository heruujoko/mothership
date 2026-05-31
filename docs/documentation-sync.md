# Documentation Sync Automation

Mothership keeps feature documentation current with a follow-up pull request
after each feature-bearing PR is merged to `main`.

## When the workflow runs

The documentation sync workflow runs when a pull request targeting `main` is
closed and merged. It skips documentation-only and automation-only changes.
Feature-bearing changes include files under:

- `agents/`
- `mothership-config/`
- `skills/`
- `tools/`
- `install.sh`

Maintainers can also run the workflow manually with `workflow_dispatch` by
providing a PR number.

## What the workflow updates

The workflow runs `scripts/sync_documentation.py`, which updates
`docs/features.md` with:

- a feature changelog entry for the merged PR
- a feature description entry
- usage help for maintainers and users

The script is idempotent. If a PR number is already documented, it exits without
modifying the documentation.

## How a documentation PR is created

When `docs/features.md` changes, the workflow creates a branch named
`docs/sync-pr-<number>`, commits the generated documentation, pushes the branch,
and opens a pull request back to `main`.

The generated PR is intentionally reviewable rather than silently committed to
`main`. If the merged feature PR did not include enough documentation source
material, maintainers should expand the generated placeholder text before
merging the documentation PR.

## Recommended PR body sections

Feature PRs should include these sections so the follow-up documentation PR is
useful immediately:

```markdown
## Feature changelog
Short user-visible release note.

## Feature description
What the feature does and why it exists.

## Usage help
Commands, setup steps, workflows, or host-specific usage notes.
```

The sync script also supports explicit extraction blocks:

```markdown
<!-- mothership-docs:changelog:start -->
Short user-visible release note.
<!-- mothership-docs:changelog:end -->

<!-- mothership-docs:description:start -->
What the feature does and why it exists.
<!-- mothership-docs:description:end -->

<!-- mothership-docs:usage:start -->
Commands, setup steps, workflows, or host-specific usage notes.
<!-- mothership-docs:usage:end -->
```
