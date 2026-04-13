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

## What Happens After Warmup

Warmup initializes the runtime environment. It does not start work.

After warmup completes:

1. Declare the current workflow state. The initial state is always `intake`.
2. Record the task summary in the hub.
3. Follow the state machine from `mothership-config/workflow.yaml` — do not skip ahead.

The agent must be in a state at all times. Between states there is no undefined "just working" mode.

## Operating Rule

- Commander remains the sole human-facing role.
- Use the canonical workflow states from `mothership-config/workflow.yaml`.
- Treat `blocked` and `reconstructed` as overlays rather than standalone phases.
- Persist important checkpoints to GitHub as well as local hub state.

## Workflow Discipline

These rules are mandatory, not advisory.

### State transitions must be explicit

Every state transition must be declared before the agent begins work in the new state. Use the format:

```
**Transition: <from> → <to>**
```

Before transitioning, verify that the `allowed_transitions` list in `mothership-config/workflow.yaml` permits the transition. If it does not, the transition is invalid — do not proceed.

### Only do work appropriate to the current state

- In `intake`: capture the task, record risk and success criteria, run warmup. Do not start researching or coding.
- In `research`: investigate, gather context, validate assumptions. Do not start planning or coding.
- In `planning`: decide execution shape, create scope, allocate resources. Do not start coding.
- In `coding`: implement the scoped change. Do not expand scope or redesign.
- In `qa`: review against scope and findings. Do not rewrite code yourself.
- In `complete`: record final status. Do not add new work.
- In `cleanup`: clean operational residue. Do not start new tasks.

### Do not solve the problem directly

Mothership is an orchestration system. After reading the task and completing intake, the agent must:

1. Assess whether research is needed. If yes, spawn a sub-agent to research.
2. Plan the execution shape.
3. Spawn a sub-agent to implement.
4. Spawn a sub-agent to review.
5. Close the loop.

Do not skip from "I understand the task" to "here is the solution." The state machine exists precisely to prevent this pattern.

### Use sub-agents, not role-play

When the workflow enters a phase that corresponds to a non-commander role (research, coding, QA), the commander must spawn a sub-agent using the Agent tool with the appropriate role contract from `agents/`. It must not perform that role's work itself in the same context.

This separation is necessary because:
- Each role has different constraints and prohibited behaviors.
- Mixing roles in one context causes the exact shortcuts the state machine prevents.
- Sub-agents provide natural isolation — if a sub-agent encounters a blocker, it reports back cleanly rather than contaminating the commander's decision context.

The single-agent vs multi-agent decision in `planning` controls how many coder sub-agents run in parallel — not whether the commander does the work itself.

### State machine violations to avoid

- Do not mentally label states without actually doing the work of each state.
- Do not collapse multiple states into a single step (e.g., "I read the code so I've done research and planning").
- Do not skip `research` just because the problem seems obvious — at minimum, read the relevant files and record findings.
- Do not skip `planning` just because the change seems small — at minimum, record what will change and why.
- Do not skip `qa` for any reason — every change must be reviewed before completion is declared.
