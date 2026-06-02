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
- discoveries from one session being invisible to the next

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

- **Requests become scoped tasks.** Intake uses brainstorming to capture success criteria, obvious dependencies, initial risk, and the first design questions before anyone starts building.
- **Research happens before planning.** Assumptions get tested with explicit research findings instead of being skipped as "obvious."
- **Planning decides execution shape.** Brainstorming and explicit plan-making happen before coding starts, so single-agent vs multi-agent work, worktree use, and issue breakdown are chosen intentionally.
- **Coding stays scoped.** Coder agents are told to execute an approved plan, not quietly redesign the roadmap mid-implementation.
- **QA is mandatory.** A task does not skip from "code exists" to "complete."
- **Cleanup closes the loop.** Branches, worktrees, dirty state, and residual artifacts are handled explicitly.
- **Knowledge compounds across sessions.** Discoveries from research, coding, and QA are captured, verified, and promoted into a persistent wiki that future tasks can read.

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

### 4. Compounding knowledge layer (wiki)

Mothership ships with an optional wiki that persists discoveries across
sessions. It follows a two-tier model inspired by the Karpathy LLM wiki
pattern — plain Markdown, no database, no vector search.

**L1 — Staging (`.draft/`)**: During a task, sub-agents append raw
discoveries to per-task draft folders. No citations needed. Lightweight and
append-only. Discarded if not promoted within 2 completed tasks.

**L2 — Verified (`insights/`, `decisions/`)**: At task completion, the
commander promotes L1 items through a quality gate — each item needs a
verifiable citation (PR URL, commit SHA, or issue number), must have survived
QA, and must describe something useful to a future session.

The directory structure lives under a configurable wiki root:

```
<wiki-root>/
  index.md                       # Master catalog (L2 only)
  log.md                         # Global timeline
  projects/<project-name>/
    index.md                     # Project catalog
    log.md                       # Project timeline
    .draft/<task-id>/            # L1 staging (per-task)
    insights/                    # L2 verified insights
    decisions/                   # L2 verified decisions
```

At the start of each task, the commander reads the project's `index.md` to
load context from previous work. Future sessions build on past discoveries
without anyone needing to re-learn what was already figured out.

Setup and protocol are documented in:
- [skills/mothership/wiki-protocol.md](skills/mothership/wiki-protocol.md) — role permissions, quality gate, promotion process
- [mothership-config/wiki-schema-default.md](mothership-config/wiki-schema-default.md) — writing conventions and templates

### 5. Mandatory review loop

The workflow does not end at implementation. QA reviews against scope, research,
and risk profile before the task can reach `complete`.

### 6. Explicit human pauses

`waiting_for_human` is a real state. That matters because many tasks fail when
automation keeps moving after it should have stopped and asked for judgment.

## Roles

Core roles are defined in [mothership-config/roles.md](mothership-config/roles.md).

Host-specific sub-agent mappings are defined in
[agents/registry.yaml](agents/registry.yaml), and the delegation contract is in
[skills/mothership/subagent-protocol.md](skills/mothership/subagent-protocol.md).

- `commander`: owns workflow, state transitions, GitHub coordination, and human communication
- `researcher`: validates assumptions, risks, constraints, and open questions
- `coder`: implements a scoped issue in a branch and optional worktree
- `qa`: verifies the implementation against issue scope and research findings
- `PR monkey`: maintains pull requests with `gh`, routes review feedback back through mothership, and updates threads after routed fixes land

The commander is the only human-facing role. It should not secretly do the work
of researcher, coder, or QA in the same context.

That separation is deliberate. It reduces hidden shortcuts and makes each phase
of work easier to inspect.

Mothership now makes that separation host-explicit:

- Claude should use the `Agent` tool family with the mapped `subagent_type`
- Codex should use `spawn_agent` with the mapped `agent_type`
- the role contract in `agents/*.md` is always part of the delegated prompt
- commander must not perform ad hoc QA after a coder returns

## Included Skills

This repository ships with:

- the installable [`mothership` skill](skills/mothership/SKILL.md) as the manual entrypoint
- the [`mothership:setup` skill](skills/mothership-setup/SKILL.md) for wiki configuration
- the [`model-policy-setup` skill](skills/model-policy-setup/SKILL.md) to bootstrap `mothership-config/model-policy.yaml`
- workflow support docs in `skills/` for intake, risk, parallelization, QA, worktrees, and escalation
- bundled `obra` skills for brainstorming, plan making, plan execution, research, and debugging flows
- focused git/GitHub skills for `commit`, `commit-push`, `create-pr`, and `pr-maintainer`

Mothership is intentionally a manual entrypoint in mixed-skill environments. It
does not assume it should take over every repository automatically.

## Installation

Run the installer from the repository root:

```bash
./install.sh
```

The root script is a wrapper around the Go installer in `tools/installer/`.
It requires a local `go` toolchain.

The script asks two questions:

1. install target: `codex`, `claude`, `antigravity`, `opencode`, or `pi`
2. install mode: `symlink` or `copy`

Installed package contents:

- `skills/`
- `agents/`
- `mothership-config/`
- `.mothership/hub/README.md`

Target roots:

- `codex` -> `${CODEX_HOME:-~/.codex}`
- `claude` -> `~/.claude`
- `antigravity` -> `~/.antigravity`
- `opencode` -> `${OPENCODE_HOME:-~/.config/opencode}`
- `pi` -> `${PI_HOME:-~/.agents}`

### Pi Setup (with team-first routing)

Pi uses `~/.agents` as its install root (not `~/.pi`) to avoid conflicts with
Codex. The `PI_HOME` environment variable overrides this.

Prerequisites:

```bash
# 1. Install pi-crew (required for team-first routing)
pi install npm:pi-crew

# 2. Install required extensions (idempotent)
pi-crew pi install npm:pi-powerline-footer
pi-crew pi install npm:@juicesharp/rpiv-todo

# 3. Run the mothership installer
./install.sh --target pi --mode copy
```

The installer copies skills, agents, config, the hub contract, and Pi runtime
extension files to `$PI_HOME/`. Verify:

```bash
ls $PI_HOME/extensions/subagent.ts
ls $PI_HOME/extensions/pi-routing.js
```

**Verifying installation:** After installation, confirm the expected directory structure:

```bash
# Verify the target directory contains expected content
ls -R $PI_HOME | grep -E '^(skills|agents|mothership-config)'

# For Pi specifically, check that extensions are installed
if [ -d "$PI_HOME/extensions" ]; then
  ls $PI_HOME/extensions/*.ts $PI_HOME/extensions/*.js 2>/dev/null || echo "Extensions directory missing"
fi
```

**Team-first delegation:** When Pi-crew is installed, mothership automatically
uses the `team` tool for role delegation. This enables:
- model routing via `mothership-config/model-policy.yaml` (role tiers + host models)
- consistent runtime state across sessions
- reduced subprocess overhead
- fallback to legacy subprocess spawning if `team` is unavailable

Non-interactive examples:

```bash
./install.sh --target claude --mode symlink
./install.sh --target codex --mode copy
./install.sh --target antigravity --mode symlink --force
./install.sh --target pi --mode copy
```

## Configuration

After installation, configure model routing (required for Pi team-first routing):

```bash
# Initialize model-policy.yaml (if not auto-created by warmup)
skills/model-policy-setup/scripts/setup-model-policy.sh

# Edit the configuration
vim mothership-config/model-policy.yaml
```

### Model Policy Configuration

The [model-policy.yaml](mothership-config/model-policy.yaml) file controls:

- **host aliases**: Map local host names to supported targets (`claude`, `codex`, `pi`, `hermes`, `ollama`)
- **role tiers**: Assign model tiers per role (`commander`, `researcher`, `coder`, `qa`, `pr_monkey`)
- **risk overrides**: Adjust tier selection based on task risk (`low`/`medium`/`high`)
- **host models**: Define available models per host and tier
- **fallback behavior**: How to handle missing/invalid config

#### Example Configuration

```yaml
# mothership-config/model-policy.yaml

# Host aliases - map local host names to supported targets
host_aliases:
  claude: claude
  codex: codex
  pi: pi
  hermes: hermes
  ollama: ollama

# Role tiers - map each role to a model tier
role_tiers:
  commander: high
  researcher: high
  coder: medium
  qa: medium
  pr_monkey: low

# Risk overrides - adjust tier based on task risk
risk_overrides:
  low:
    coder: low
    qa: low
  medium: {}
  high:
    researcher: high
    coder: high
    qa: high

# Host models - available models per host and tier
host_models:
  pi:
    high: ["claude-3-5-sonnet-latest", "claude-3-5-haiku-latest"]
    medium: ["claude-3-5-sonnet-latest", "claude-3-5-haiku-latest"]
    low: ["claude-3-5-haiku-latest"]
  codex:
    high: ["claude-3-5-sonnet-latest"]
    medium: ["claude-3-5-haiku-latest"]
    low: ["claude-3-5-haiku-latest"]
```

Model resolution order:
1. Role tier from `role_tiers` + risk override from `risk_overrides` in this file
2. Host-specific model mapping in `host_models`
3. Fallback: inherit from current session model (legacy behavior)

## Using It

Invoke the `mothership` skill when you want the orchestration workflow itself,
not just isolated code help.

High-level flow:

1. Start with `mothership`.
2. Run the intake warm-up from `skills/mothership/scripts/warmup.sh`.
   - Warm-up warns (does not hard-fail) when `mothership-config/model-policy.yaml` is missing or invalid, and falls back to legacy model inheritance.
3. Read the canonical config:
   - `mothership-config/workflow.yaml`
   - `mothership-config/roles.md`
   - `mothership-config/model-policy.yaml`
   - `agents/registry.yaml`
   - `mothership-config/skill-dependencies.md`
   - `skills/mothership/subagent-protocol.md`
4. Record the task in the local hub.
5. Move through the workflow state machine explicitly.
6. Persist meaningful checkpoints to GitHub.
7. Close the loop with QA and cleanup.

### Setting up the wiki (optional)

Run `/mothership:setup` once per project to configure the wiki. The setup
script (`skills/mothership/scripts/wiki-setup.sh`) will:

1. Ask for a wiki root location (default: `~/.mothership/wiki/`)
2. Create the directory structure under `projects/<project-name>/`
3. Write preferences to `.mothership/wiki.yaml`
4. Initialize `schema.md`, `index.md`, and `log.md`

After setup, wiki operations happen automatically as side-effects of the
workflow — no extra steps needed. Sub-agents write to L1 staging during
research, coding, and QA. The commander promotes verified items to L2 at
completion.

When no wiki is configured (no `.mothership/wiki.yaml`), all wiki operations
are silently skipped.

Required auxiliary skills are listed in
[mothership-config/skill-dependencies.md](mothership-config/skill-dependencies.md).

## Repository Layout

- `agents/`: role contracts for commander, researcher, coder, QA, and PR monkey
- `agents/registry.yaml`: host-specific mapping from mothership roles to the sub-agent tool flavor
- `mothership-config/`: canonical workflow, roles registry, skill dependencies, and wiki schema
- `skills/`: installable skills plus supporting workflow guidance
- `skills/mothership/wiki-protocol.md`: wiki role permissions, quality gate, and promotion process
- `skills/mothership-setup/SKILL.md`: wiki setup skill (invoked as `/mothership:setup`)
- `skills/mothership/scripts/wiki-setup.sh`: wiki initialization script
- `.mothership/hub/`: local hub contract and runtime state pattern
- `.mothership/`: ignored runtime artifacts created during execution

## Design Principles

- **State over vibes.** Every task is in a defined state.
- **Durability over convenience.** Important decisions should survive the session.
- **Knowledge should compound.** Discoveries from past sessions should be available to future ones.
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
