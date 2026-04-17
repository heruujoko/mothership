# Mothership

Mothership is a workflow package for running engineering work with explicit
state, durable checkpoints, and clean handoffs between orchestration,
implementation, and review.

It is designed for teams that want AI-assisted execution without the usual
failure modes:

- scope drift hidden inside a long session
- coding before requirements are clarified
- "done" being declared without real QA
- task context disappearing between iterations
- local agent memory diverging from GitHub reality

Mothership treats software work like an operational system, not a chat thread.
Every task has a state. Every transition is explicit. Important decisions are
written down. Review is mandatory. Cleanup is part of the job.

## What Mothership Is

Mothership is not a single super-agent.

It is a structured orchestration model built around:

- a `commander` role that owns the workflow and talks to humans
- specialized roles for research, coding, and QA
- a canonical state machine in [mothership-config/workflow.yaml](mothership-config/workflow.yaml)
- a local hub under `.mothership/` for live operational memory
- GitHub artifacts as the durable external record

The result is a more maintainable way to run engineering work:

- tasks stay auditable
- partial progress survives interruptions
- engineers can see what state a task is in
- QA has a defined place in the loop
- cleanup and branch hygiene are treated as first-class work

## Why Engineers Use It

Most engineering mess does not come from writing code. It comes from losing the
shape of the work.

Mothership helps engineers maintain their work by making the execution model
visible:

- **Requests become scoped tasks.** Intake captures success criteria, obvious dependencies, and initial risk before anyone starts building.
- **Research happens before planning.** Assumptions get tested before implementation choices harden.
- **Planning decides execution shape.** Single-agent vs multi-agent work, worktree use, and issue breakdown are chosen intentionally.
- **Coding stays scoped.** Coder agents are told to implement the issue, not quietly redesign the roadmap.
- **QA is mandatory.** A task does not skip from "code exists" to "complete."
- **Cleanup closes the loop.** Branches, worktrees, dirty state, and residual artifacts are handled explicitly.

This is especially useful for:

- repositories with frequent interruptions and task switching
- teams using issues and PRs as the source of truth
- engineering leads who want clearer audit trails
- solo engineers who want a repeatable system instead of ad hoc agent sessions

## The Workflow

The workflow is defined in
[mothership-config/workflow.yaml](mothership-config/workflow.yaml).

Canonical states:

1. `intake`
2. `research`
3. `planning`
4. `coding`
5. `qa`
6. `waiting_for_human`
7. `complete`
8. `cleanup`

Two overlays can apply on top of any active state:

- `blocked`
- `reconstructed`

That distinction matters. `blocked` is not a replacement for the current phase.
It means the task is still, for example, in `coding`, but cannot continue until
the blocker is resolved.

### State Intent

| State | What it protects |
| --- | --- |
| `intake` | Starting with real task context, warm-up hygiene, and risk visibility |
| `research` | Validating assumptions before planning or coding |
| `planning` | Choosing scope, isolation, and execution strategy deliberately |
| `coding` | Keeping implementation tied to a defined issue and branch |
| `qa` | Verifying correctness instead of assuming effort equals quality |
| `waiting_for_human` | Pausing honestly when only a human can decide or verify |
| `complete` | Marking success only after evidence and references are in place |
| `cleanup` | Removing operational residue before the task is truly closed |

### Exit Criteria Are the Point

Each state in `workflow.yaml` has exit criteria. Mothership uses those criteria
to prevent the most common agent shortcut: mentally skipping ahead.

Examples:

- `intake` is not done until warm-up runs, `.mothership/` hygiene is enforced,
  auxiliary skills are checked, and the task summary plus initial risk are recorded.
- `coding` is not done until a PR exists or an implementation checkpoint is recorded.
- `qa` is not done until the disposition is explicit.
- `complete` is not done until GitHub status and references are current.
- `cleanup` is not done until worktree and dirty-state handling are recorded.

That discipline is what makes the workflow useful in real repos over time.

## How Mothership Maintains Work

Mothership is built around the idea that engineering work should remain
recoverable, inspectable, and resumable.

### 1. Local operational memory

The local hub in [.mothership/hub/README.md](.mothership/hub/README.md) tracks
the live state of the task:

- active task
- issues and PRs
- agent assignments
- worktrees
- escalations
- cleanup status

This gives the workflow a working memory that is separate from any one agent
session.

### 2. GitHub as durable record

Important checkpoints are expected to be written back to GitHub. That keeps the
task understandable even if local state is lost or a session is interrupted.

### 3. Reconstruction when things go wrong

If hub state is corrupted or incomplete, Mothership can mark the task as
`reconstructed` and rebuild state from GitHub artifacts rather than pretending
nothing happened.

### 4. Mandatory review loop

The workflow does not end at implementation. QA reviews against scope, research,
and risk profile before the task can reach `complete`.

### 5. Explicit human pauses

`waiting_for_human` is a real state. That matters because many tasks fail when
automation keeps moving after it should have stopped and asked for judgment.

## Roles

Core roles are defined in [mothership-config/roles.md](mothership-config/roles.md).

- `commander`: owns workflow, state transitions, GitHub coordination, and human communication
- `researcher`: validates assumptions, risks, constraints, and open questions
- `coder`: implements a scoped issue in a branch and optional worktree
- `qa`: verifies the implementation against issue scope and research findings
- `PR monkey`: maintains pull requests with `gh`, routes review feedback back through mothership, and updates threads after routed fixes land

The commander is the only human-facing role. It should not secretly do the work
of researcher, coder, or QA in the same context.

That separation is deliberate. It reduces hidden shortcuts and makes each phase
of work easier to inspect.

## Included Skills

This repository ships with:

- the installable [`mothership` skill](skills/mothership/SKILL.md) as the manual entrypoint
- workflow support docs in `skills/` for intake, risk, parallelization, QA, worktrees, and escalation
- bundled `obra` skills used during research, planning, and debugging flows
- focused git/GitHub skills for `commit`, `commit-push`, `create-pr`, and `pr-maintainer`

Mothership is intentionally a manual entrypoint in mixed-skill environments. It
does not assume it should take over every repository automatically.

## Using It

Invoke the `mothership` skill when you want the orchestration workflow itself,
not just isolated code help.

High-level flow:

1. Start with `mothership`.
2. Run the intake warm-up from `skills/mothership/scripts/warmup.sh`.
3. Read the canonical config:
   - `mothership-config/workflow.yaml`
   - `mothership-config/roles.md`
   - `mothership-config/skill-dependencies.md`
4. Record the task in the local hub.
5. Move through the workflow state machine explicitly.
6. Persist meaningful checkpoints to GitHub.
7. Close the loop with QA and cleanup.

Required auxiliary skills are listed in
[mothership-config/skill-dependencies.md](mothership-config/skill-dependencies.md).

## Repository Layout

- `agents/`: role contracts for commander, researcher, coder, QA, and PR monkey
- `mothership-config/`: canonical workflow, roles registry, and skill dependencies
- `skills/`: installable skills plus supporting workflow guidance
- `.mothership/hub/`: local hub contract and runtime state pattern
- `.mothership/`: ignored runtime artifacts created during execution

## Design Principles

- **State over vibes.** Every task is in a defined state.
- **Durability over convenience.** Important decisions should survive the session.
- **Research before implementation when risk is real.**
- **QA is part of delivery, not an optional afterthought.**
- **Blocked work should be visible, not buried.**
- **Cleanup is part of completion.**

## Who This Is For

Mothership is a good fit if you want AI-assisted engineering work to feel more
like a disciplined production system and less like a sequence of clever prompts.

It is for teams and engineers who care about:

- maintaining momentum across interruptions
- keeping issues, PRs, and local execution aligned
- reducing ambiguous ownership between planning, coding, and review
- making AI-driven work easier to trust, audit, and resume

If that is the bar you want, Mothership gives you a concrete workflow for
reaching it.
