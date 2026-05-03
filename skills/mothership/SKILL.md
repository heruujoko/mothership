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
5. Ensure `agents/registry.yaml` and `skills/mothership/subagent-protocol.md` are available.
6. If preflight fails, record a `blocked` overlay or explicit fallback notes before continuing.

## Purpose

Run the mothership workflow package in a realistic mixed-skill environment where
other unrelated skills may also be available.

When this skill is invoked, use the local package artifacts in this order:

1. Read `mothership-config/roles.md` for the available role contracts.
2. Read `mothership-config/workflow.yaml` for the canonical workflow states and transitions.
3. Read `agents/registry.yaml` for host-specific sub-agent mappings.
4. Read `skills/mothership/subagent-protocol.md` for the host-aware delegation contract.
5. Read `.mothership/hub/README.md` for the local hub state contract.
6. Read `mothership-config/skill-dependencies.md` for required external skills.
7. Read the supporting workflow docs in `skills/` as needed:
   - `task-intake-and-decomposition/SKILL.md`
   - `risk-assessment/SKILL.md`
   - `research-execution/SKILL.md`
   - `parallelization-decision/SKILL.md`
   - `github-issue-management/SKILL.md`
   - `pr-maintainer/SKILL.md`
   - `worktree-management/SKILL.md`
   - `qa-review-loop/SKILL.md`
   - `human-escalation/SKILL.md`
8. Apply the role docs in `agents/` when acting within a specific mothership role.

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

When the workflow enters a phase that corresponds to a non-commander role (research, coding, QA), the commander must spawn a sub-agent using the host mapping from `agents/registry.yaml` and the role contract from `agents/`. It must not perform that role's work itself in the same context.

This separation is necessary because:
- Each role has different constraints and prohibited behaviors.
- Mixing roles in one context causes the exact shortcuts the state machine prevents.
- Sub-agents provide natural isolation — if a sub-agent encounters a blocker, it reports back cleanly rather than contaminating the commander's decision context.

The single-agent vs multi-agent decision in `planning` controls how many coder sub-agents run in parallel — not whether the commander does the work itself.

### Host-aware sub-agent protocol

Use `skills/mothership/subagent-protocol.md` as the execution contract for all
delegation.

Required delegation behavior:

1. Identify the active workflow state.
2. Map the state to a role using `agents/registry.yaml`.
3. Use the host-specific sub-agent tool declared for that role.
4. Build the prompt from the full role contract plus task-specific context.
5. Wait for the delegated result before doing work that belongs to that role.
6. Record the output and only then decide the next transition.

For Claude:

- Use the `Agent` tool family and the `subagent_type` declared in `agents/registry.yaml`.
- `researcher` must use `Explore`.
- `coder` must use `general-purpose`.
- `qa` must use `feature-dev:code-reviewer`.

For Codex:

- Use `spawn_agent` and the `agent_type` declared in `agents/registry.yaml`.
- `researcher` should use `explorer`.
- `coder` should use `worker`.
- `qa` should use `explorer`.

The commander may gather minimal routing context in `intake` and `planning`.
After delegation starts, it must not continue the delegated work itself.

Examples of prohibited drift:

- "Let me verify the results independently before marking QA complete."
- "I'll quickly inspect the changed files myself instead of spawning QA."
- "The coder already succeeded, so I can close the loop without review."

### PR work uses PR monkey

When the user provides a PR, asks to create a PR from the current project, or
asks to process PR review feedback, commander should spawn `agents/pr-monkey.md`
for the PR-facing work using the host mapping from `agents/registry.yaml`.

PR monkey is responsible for:

- making sure `gh` is installed and authenticated before PR work proceeds
- creating PRs with full descriptions
- reading PR feedback one item at a time
- returning each actionable item to commander as new scoped work
- pushing follow-up fixes, replying to comments, and resolving threads after the
  routed fix is complete

PR monkey must never implement feedback directly outside the main workflow and
must never merge without explicit human consent.

### State machine violations to avoid

- Do not mentally label states without actually doing the work of each state.
- Do not collapse multiple states into a single step (e.g., "I read the code so I've done research and planning").
- Do not skip `research` just because the problem seems obvious — at minimum, read the relevant files and record findings.
- Do not skip `planning` just because the change seems small — at minimum, record what will change and why.
- Do not skip `qa` for any reason — every change must be reviewed before completion is declared.
