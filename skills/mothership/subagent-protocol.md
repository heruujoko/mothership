# Sub-Agent Protocol

This document makes mothership delegation behavior explicit across hosts.

Use it together with:

- `mothership-config/workflow.yaml`
- `mothership-config/roles.md`
- `agents/registry.yaml`
- the role contract in `agents/<role>.md`

## Core Rule

Commander orchestrates. It does not perform research, coding, QA, or PR
maintenance itself once the task enters a delegated state.

Minimal routing reads are allowed during `intake` and `planning`. After a
delegated role is spawned, commander may only:

- summarize the returned output
- decide the next state transition
- record hub or GitHub checkpoints
- escalate blockers to the human
- spawn the next role

Commander must not:

- continue the delegated task itself "just to verify quickly"
- inspect files or run tests in a way that substitutes for the delegated role
- perform QA after a coder returns instead of spawning QA
- resolve PR feedback itself when `pr_monkey` is the active delegated role

If commander says things like "let me verify independently" after coding, the
workflow has drifted out of compliance.

## Role Selection

Use `agents/registry.yaml` as the single source of truth for host mapping.

Delegated states map to roles as follows:

- `research` -> `researcher`
- `coding` -> `coder`
- `qa` -> `qa`
- explicit PR creation or PR feedback work -> `pr_monkey`

`planning`, `complete`, and `cleanup` remain commander-owned states.

## Model Resolution

Before spawning a delegated role, resolve model selection with `mothership-config/model-policy.yaml`:

1. explicit per-invocation override (if provided)
2. `risk_overrides` (if task risk is known)
3. `role_tiers` mapping for delegated role
4. host tier default using `host_aliases` + `host_models`/`tiers`
5. fallback: current thread model (legacy inheritance)

If policy is missing or invalid, warn and continue with fallback instead of hard-failing.

## Prompt Assembly

When spawning a sub-agent, construct the prompt from these parts in order:

1. Full role contract from the mapped `agents/*.md` file.
2. Current workflow state and the transition that entered it.
3. Task summary, success criteria, and risk notes.
4. Relevant issue, PR, branch, worktree, or file context.
5. Explicit write scope and ownership boundaries when the role can modify files.
6. Required output format for the role.
7. Any blockers, constraints, or required follow-up routing.

Do not replace the role contract with a short paraphrase. The contract is part
of the instruction set, not optional flavor text.

## Required Output Shapes

### Researcher

Return:

- summary
- assumptions
- constraints
- risks
- recommended approach
- open questions
- implementation considerations

When wiki is configured, the researcher should also append discoveries to
`.draft/<task-id>/insights.md` using the format:
`### [research] <summary>` followed by the finding.
This is in addition to (not instead of) the structured research output above.

### Coder

Return:

- implemented scope
- files changed
- verification performed
- blockers or follow-up items
- PR or branch status when applicable

When wiki is configured, the coder should also append implementation discoveries
to `.draft/<task-id>/insights.md` using the format:
`### [coding] <summary>` followed by the finding.
This includes non-obvious gotchas, patterns discovered, and architectural observations.

### QA

Return:

- scope coverage
- findings by severity
- verification evidence
- required changes
- final disposition: approved / changes requested / escalate

When wiki is configured, QA should note which L1 items from `.draft/<task-id>/`
were confirmed or challenged in `.draft/<task-id>/notes.md`:
- `### [qa] confirmed: <item summary>` for items that survived review
- `### [qa] challenged: <item summary>` for items that are incorrect or superseded

### PR Monkey

Return:

- PR status or URL
- current feedback item being handled
- commands or replies posted
- blockers preventing continuation

## Host Mapping

### Claude Code

Use the `Agent` tool family with the `subagent_type` declared in
`agents/registry.yaml`.

Examples:

- `researcher` -> `Explore`
- `coder` -> `general-purpose`
- `qa` -> `feature-dev:code-reviewer`
- `pr_monkey` -> `general-purpose`

Pass the rendered role prompt as the task prompt and keep the description short
and operational.

### Codex

Use `spawn_agent` with the `agent_type` declared in `agents/registry.yaml`.

Examples:

- `researcher` -> `explorer`
- `coder` -> `worker`
- `qa` -> `explorer`
- `pr_monkey` -> `default`

When the delegated task depends on current thread context, fork the context.
For coder roles, assign clear file or module ownership in the prompt.

### Pi

Pi delegation is team-first when the runtime exposes the `team` tool.

`mothership_spawn` reads `agents/registry.yaml` (`roles.<role>.pi`) and:

1. Tries direct `team` tool invocation with mapped `team`, optional
   `model_hint`, and optional `model_fallback_chain` when
   `team_path_enabled: true` and the configured `team_tool` is available.
   For safety, `team_tool` must be exactly `team` unless explicitly sanctioned
   via `team_tool_allow_unsafe: true`.
2. Falls back to legacy subprocess spawning (`pi --mode json`) when team-path
   is disabled, unmapped, unavailable, or unhealthy.

Fallback reasons are logged explicitly (for example:
`team_path_disabled`, `team_mapping_missing`, `team_tool_unsafe`,
`team_tool_unavailable`, `team_tool_unhealthy`).

Legacy path arguments remain:

- `role` — the mothership role to invoke (`researcher`, `coder`, `qa`, `pr_monkey`)
- `task` — the task description or question for the sub-agent
- `contract_file` — path to the role contract (`agents/<role>.md`)
- `context` — any additional context to pass (issue, branch, files, etc.)

Commander waits for completion before transitioning.

**Requirements:**
- The `mothership_spawn` runtime extension files must be installed under
  `$PI_HOME/extensions/` (default `~/.agents/extensions/`):
  `subagent.ts` and `pi-routing.js`. See `skills/mothership/SKILL.md` warmup checklist.
- The install target for Pi is `~/.agents` (not `~/.codex` or `~/.pi`) to avoid
  conflicts with other agent hosts. Set the `PI_HOME` environment variable to
  override.
- **Required prerequisite:** `pi install npm:pi-crew`
- **Required extension installs** (idempotent):
  ```bash
  pi-crew pi install npm:pi-powerline-footer
  pi-crew pi install npm:@juicesharp/rpiv-todo
  ```
- The subprocess inherits the pi working directory. File paths in the context
  must be absolute or relative to the working directory.
- Model should follow policy resolution first; if policy is missing/invalid, subprocess inherits the current thread model unless overridden in extension config.

**Secure config guidance:**
- Do not commit secrets or API keys to agent config files under `$PI_HOME/`.
- Use user-scope configuration only. `$PI_HOME/` is user-local and should not
  be shared across users or committed to version control.
- If model fallback chains reference provider keys, store them in environment
  variables or the provider's native credential store — never in YAML config.

## State Gates

Before leaving a delegated state, commander must have evidence that:

- the delegated sub-agent was actually invoked for that state
- the delegated sub-agent returned a result or blocker
- the result was recorded in `state.yaml` under the relevant phase section and in GitHub, with only a reference stored inline

This is especially strict for `qa`: commander cannot mark QA complete based on
its own ad hoc verification after a coder returns.

## Hub State Embedding

Sub-agent output is embedded directly in `state.yaml` under the phase key
(e.g., `research:`, `coding:`). There are no separate ref files — the entire
session state is one file.

1. Write the output inline to `state.yaml` under the appropriate phase section
2. Reference important GitHub artifacts (issue comment, PR description) in the
   phase section if needed for durability

This keeps the full session state in one read (~400-800 tokens).
