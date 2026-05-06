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

Pi has no built-in sub-agent spawning. Delegation requires the
`mothership_spawn` extension tool to be loaded (installed under
`~/.pi/agent/extensions/` alongside the mothership package).

Use `mothership_spawn` with the arguments declared in `agents/registry.yaml`
under the `pi` host for each role:

- `role` — the mothership role to invoke (`researcher`, `coder`, `qa`, `pr_monkey`)
- `task` — the task description or question for the sub-agent
- `contract_file` — path to the role contract (`agents/<role>.md`)
- `context` — any additional context to pass (issue, branch, files, etc.)

The extension spawns a new `pi` subprocess for the delegated role and returns
its output to commander. Commander waits for completion before transitioning.

**Requirements:**
- The `mothership_spawn` extension must be installed. See `skills/mothership/SKILL.md`
  warmup checklist.
- The subprocess inherits the pi working directory. File paths in the context
  must be absolute or relative to the working directory.
- The subprocess uses the same model unless overridden in the extension config.

## State Gates

Before leaving a delegated state, commander must have evidence that:

- the delegated sub-agent was actually invoked for that state
- the delegated sub-agent returned a result or blocker
- the result was recorded in hub state, GitHub, or both

This is especially strict for `qa`: commander cannot mark QA complete based on
its own ad hoc verification after a coder returns.
