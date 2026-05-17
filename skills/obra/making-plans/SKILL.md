---
name: making-plans
description: Bundled dependency skill used by mothership intake, research, and planning phases to turn explored requirements into an explicit implementation plan before coding starts.
---

# obra/making-plans

Use this skill after `obra/brainstorming` has clarified the task and the
relevant research has been gathered.

## Intended Usage

- turn clarified requirements into an explicit implementation plan
- sequence the work into ordered, testable steps
- record dependencies, blockers, and validation points before coding starts
- give coder agents a concrete plan that `obra/executing-plans` can follow

## Mothership Routing

- In `research`, use this to package findings into a planning-ready handoff.
- In `planning`, use this to create the approved implementation plan before
  transitioning to `coding`.
- Do NOT use this skill to implement code directly.

## Note

In environments with an upstream obra package, this local marker may be
replaced or shadowed by the upstream installation.
