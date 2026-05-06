---
name: systematical-debugging
description: Bundled dependency skill used by mothership coding after QA feedback for structured debugging.
---

# obra/systematical-debugging

This bundled skill marker exists so mothership can preflight and invoke
`obra/systematical-debugging` when coding receives QA-requested changes.

## Intended Usage

- reproduce reported failures
- isolate likely root causes
- verify fixes before resubmission

## Note

In environments with an upstream obra package, this local marker may be replaced
or shadowed by the upstream installation.
