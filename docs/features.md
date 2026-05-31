# Feature Documentation

This page is the canonical feature documentation index for Mothership. It is
updated after feature-bearing pull requests merge to `main`, then refined in the
follow-up documentation PR when a generated entry needs more detail.

## What each feature entry must include

Every documented feature should include:

- a changelog entry that identifies the merged PR and the user-visible change
- a feature description that explains the capability and why it exists
- usage help that tells maintainers how to install, configure, or invoke it

## Authoring source material in feature PRs

The documentation sync job can produce better follow-up PRs when feature PRs
include one or more of these sections in the PR body:

```markdown
## Feature changelog
Short user-visible release note.

## Feature description
What the feature does and why it exists.

## Usage help
Commands, setup steps, workflows, or host-specific usage notes.
```

For more precise extraction, maintainers can use these HTML comment blocks in a
PR body:

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

If a feature PR does not include documentation source material, the generated
follow-up documentation PR contains a placeholder that must be expanded before
it is merged.

## Feature Changelog

<!-- mothership-docs:changelog:start -->
<!-- mothership-docs:changelog:end -->

## Feature Catalog

<!-- mothership-docs:features:start -->
<!-- mothership-docs:features:end -->
