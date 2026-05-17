---
name: creating-plan
description: Bundled dependency skill used by mothership research for structured plan creation.
---

# obra/creating-plan

This skill name is kept as a compatibility marker for older references.
Active mothership workflows should use `obra/making-plans` instead.

## Intended Usage

- compatibility fallback for environments that still reference `obra/creating-plan`
- redirect active planning flows to `obra/making-plans`
- avoid hard failures while older docs are updated

## Note

In environments with an upstream obra package, this local marker may be replaced
or shadowed by the upstream installation.
