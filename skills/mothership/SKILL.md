---
name: mothership
description: Use when the user explicitly invokes /mothership or asks to run the mothership orchestration workflow manually in a repository that may also have many other installed skills.
---

# Mothership

This skill is the explicit front door for the mothership package.

Use it only when the user intentionally wants to activate mothership
orchestration. Do not assume this package should take over automatically just
because it is installed.

This skill expects the full mothership package layout to be installed so it can
read `agents/`, `mothership-config/`, `.mothership/hub/`, and the supporting workflow docs.

## Startup Warm-Up Checklist

Run this checklist immediately when `mothership` is invoked:

1. Run `skills/mothership/scripts/warmup.sh` from repository root.
2. Ensure runtime artifacts are dot-prefixed under `.mothership/`.
3. Ensure `.mothership/` is present in `.gitignore`.
4. Ensure required auxiliary skills are available per `mothership-config/skill-dependencies.md`.
5. If preflight fails, record a `blocked` overlay or explicit fallback notes before continuing.

## Purpose

Run the mothership workflow package in a realistic mixed-skill environment where
other unrelated skills may also be available.

When this skill is invoked, use the local package artifacts in this order:

1. Read `mothership-config/roles.md` for the available role contracts.
2. Read `mothership-config/workflow.yaml` for the canonical workflow states and transitions.
3. Read `.mothership/hub/README.md` for the local hub state contract.
4. Read `mothership-config/skill-dependencies.md` for required external skills.
5. Read the supporting workflow docs in `skills/` as needed:
   - `task-intake-and-decomposition.md`
   - `risk-assessment.md`
   - `research-execution.md`
   - `parallelization-decision.md`
   - `github-issue-management.md`
   - `worktree-management.md`
   - `qa-review-loop.md`
   - `human-escalation.md`
6. Apply the role docs in `agents/` when acting within a specific mothership role.

## Activation Rule

- Treat `mothership` as a manual opt-in entrypoint.
- Prefer it when the user says `/mothership` or otherwise explicitly asks to run mothership workflow.
- Do not let unrelated installed skills override this entrypoint once mothership is explicitly invoked.

## Operating Rule

- Commander remains the sole human-facing role.
- Use the canonical workflow states from `mothership-config/workflow.yaml`.
- Treat `blocked` and `reconstructed` as overlays rather than standalone phases.
- Persist important checkpoints to GitHub as well as local hub state.
