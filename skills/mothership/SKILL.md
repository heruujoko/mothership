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
6. If `mothership-config/model-policy.yaml` is missing, run `skills/model-policy-setup/scripts/setup-model-policy.sh` to bootstrap it. Inform the user they should edit the `host_models` section to match their available models.
7. If preflight fails, record a `blocked` overlay or explicit fallback notes before continuing.
8. If `.mothership/wiki.yaml` exists, read `projects/<name>/index.md` from the configured wiki root for project context.
9. On pi: verify the `mothership_spawn` extension is available under `$PI_HOME/extensions/` (default `~/.agents/extensions/`). If absent, delegation will not work — record a `blocked` overlay and continue without sub-agent spawning.

## Pi Setup Prerequisites

Before using mothership with Pi, ensure the following prerequisites are met:

1. **Install target is `~/.agents`.** The canonical Pi install root is `~/.agents` to
   avoid conflicts with Codex (which uses `~/.codex`). The installer sets
   `PI_HOME=~/.agents` so Pi resolves extensions from `$PI_HOME/extensions/`.

2. **Install pi-crew.** Mothership team-first routing requires pi-crew:
   ```bash
   pi install npm:pi-crew
   ```

3. **Install required extensions** (idempotent, safe to re-run):
   ```bash
   pi-crew pi install npm:pi-powerline-footer
   pi-crew pi install npm:@juicesharp/rpiv-todo
   ```

4. **Run the mothership installer for Pi:**
   ```bash
   ./install.sh --target pi --mode copy
   ```
   This copies skills, agents, config, and the `mothership_spawn` extension to
   `$PI_HOME/` (default `~/.agents/`).

5. **Verify the extension resolves:**
   ```bash
   ls $PI_HOME/extensions/subagent.ts
   ls $PI_HOME/extensions/pi-routing.js
   ```
   If the file is absent, delegation will not work.

6. **Secure config guidance:** Do not commit secrets or API keys to agent config
   files. Use user-scope config only (`$PI_HOME/` is user-local).

## Model Policy Configuration

Mothership routes models per role and risk level through `mothership-config/model-policy.yaml`.

### First-time setup

If the file is missing, bootstrap defaults:
```bash
skills/model-policy-setup/scripts/setup-model-policy.sh
```
This is non-destructive — it skips if the file already exists.

### Where to edit models for your crews

Edit `mothership-config/model-policy.yaml`:

```yaml
# Which model each host uses for each tier
host_models:
  pi:                          # <-- YOUR HOST
    high: "openai-codex/gpt-5.5"    # <-- commander, researcher
    medium: "openai-codex/gpt-5.3-codex"  # <-- coder, qa
    low: "openai-codex/gpt-5-mini"   # <-- pr_monkey, low-risk downgrade

# Which tier each role defaults to
role_tiers:
  commander: high     # change to "medium" to downgrade planning
  researcher: high
  coder: medium       # change to "high" for complex implementation
  qa: medium
  pr_monkey: low

# Risk-based overrides (applied before role_tiers)
risk_overrides:
  low:
    coder: low        # low-risk coding uses cheap model
    qa: low
  high:
    coder: high       # high-risk coding uses best model
    researcher: high
```

Key sections:
- **`host_models.<host>`** — map tier names to actual model IDs your provider supports
- **`role_tiers`** — change which tier a role uses by default
- **`risk_overrides`** — upgrade or downgrade based on task risk assessment
- **`fallback`** — what happens when policy is missing or invalid (default: warn + inherit current thread model)

Resolution order: explicit_override → risk_override → role_tier → host_tier_default → inherit_current_thread_model

After editing, validate:
```bash
skills/mothership/scripts/validate-model-policy.sh
```

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
3. Use `obra/brainstorming` during `intake` before any delegation.
4. Follow the state machine from `mothership-config/workflow.yaml` — do not skip ahead.
5. If wiki is configured, L1 staging writes happen automatically during research, planning, coding, and QA states.
6. L2 promotion happens automatically at `complete`.

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
- In `research`: investigate, gather context, validate assumptions, and return a planning-ready handoff. Do not start coding.
- In `planning`: present the design, create the explicit implementation plan, and allocate resources. Do not start coding until the plan exists.
- In `coding`: implement the scoped change by executing the approved plan. Do not expand scope or redesign.
- In `qa`: review against scope and findings. Do not rewrite code yourself.
- In `complete`: record final status. Do not add new work.
- In `cleanup`: clean operational residue. Do not start new tasks.

### Phase skill routing

- In `intake`: invoke `task-intake-and-decomposition` and `obra/brainstorming`.
- In `research`: the researcher must invoke `obra/brainstorming` in research mode, `obra/superpowers`, and `obra/making-plans`.
- In `planning`: invoke `obra/brainstorming` to present the design and `obra/making-plans` to record the implementation plan.
- In `coding`: the coder must invoke `obra/executing-plans`. If QA sends the work back, add `obra/systematical-debugging` before resuming.

### Do not solve the problem directly

Mothership is an orchestration system. After reading the task and completing intake, the agent must:

1. Complete `intake` with brainstorming-driven clarification.
2. Spawn a sub-agent to research.
3. Complete `planning` with a recorded design and explicit implementation plan.
4. Spawn a sub-agent to implement against that plan.
5. Spawn a sub-agent to review.
6. Close the loop.

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

For Pi:

- Use `mothership_spawn` with the `role` and `task` arguments declared in `agents/registry.yaml` under the `pi` host.
- `mothership_spawn` is team-first: it calls the configured `team` tool directly when available (`team_path_enabled`, mapped `team`, optional `model_hint`/`model_fallback_chain`). By default `team_tool` must be exactly `team` unless explicitly sanctioned (`team_tool_allow_unsafe: true`), otherwise it falls back to legacy subprocess spawning with explicit reason logging.
- See `skills/mothership/subagent-protocol.md` for the full Pi delegation protocol.

The commander may gather minimal routing context in `intake` and `planning`.
After delegation starts, it must not continue the delegated work itself.

Examples of prohibited drift:

- "Let me verify the results independently before marking QA complete."
- "I'll quickly inspect the changed files myself instead of spawning QA."
- "The coder already succeeded, so I can close the loop without review."

### PR work uses PR monkey

When the user provides a PR, asks to create a PR from the current project, or
asks to process PR review feedback, commander should spawn `agents/pr_monkey.md`
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
- Do not skip `planning` just because the change seems small — at minimum, present the design and record the implementation plan.
- Do not skip `qa` for any reason — every change must be reviewed before completion is declared.
