# Mothership Core Design

Date: 2026-04-11  
Status: Draft v1  
Scope: Brainstormed design only, no implementation

## 1. Purpose

Mothership is a local-first orchestration repository for coordinating multiple specialized agents to complete software delivery tasks. It is designed for a human operator working from a local CLI or desktop coding assistant, with GitHub serving as the durable audit trail.

The system is optimized for coding, design, and research workflows where one central commander agent decomposes work, assigns specialized agents, coordinates review, and manages completion.

## 2. Goals

- Accept a task from a human through CLI or local coding assistant.
- Analyze complexity, risk, and execution strategy before coding begins.
- Use research agents to validate assumptions and gather context.
- Use GitHub issues and PRs as durable records of task progress.
- Use isolated branches and worktrees for coder agents when appropriate.
- Use QA agents to review outputs and enforce completion quality.
- Escalate to humans when verification or ambiguity resolution is required.
- Clean up branches and worktrees after successful completion.

## 3. Non-Goals for v1

- Full autonomous project management across multiple repositories.
- Provider-agnostic model orchestration with dynamic routing.
- Rich web dashboard.
- Automatic merge to default branch without explicit human approval.
- Highly dynamic role plugin marketplace.

## 4. Core Principles

### Local-first orchestration
The commander runs locally on a Mac or VPS. Live task state is stored in local hub files.

### GitHub as durable audit
GitHub issues, comments, labels, and PRs are the persistent external record. If the local hub becomes incomplete or corrupted, the commander must reconstruct enough state from GitHub.

### Single source of truth for roles
Available roles are defined in one Markdown registry file. New roles are introduced by updating that file and adding corresponding role docs.

### Isolation by default
Coder work should prefer isolated branches and worktrees when tasks are parallelized.

### Parallelism is conditional
Commander chooses between single-agent and multi-agent execution based on merge conflict risk, task decomposition quality, and machine resources such as CPU, memory, and disk.

### Human-in-the-loop
Any ambiguity, risky change, or unverifiable requirement may trigger human escalation.

## 5. Main Roles

- Commander
- Research Agent
- Coder Agent
- QA Agent

Future roles may include:
- Designer
- Security Reviewer
- Docs Writer
- DevOps Agent

## 6. Core Architecture

```text
Human (CLI / Codex / Claude Code)
        |
        v
  Commander Agent
        |
        +-- Local Hub State
        +-- Role Registry
        +-- GitHub Issues / PRs / Comments
        +-- Worktree Manager
        +-- Agent Spawner
                |
                +-- Research Agents
                +-- Coder Agents
                +-- QA Agents
```

## 7. Durable State Model

### Local hub
The local hub is the operational state store. It tracks:
- root task id
- task summary
- risk level
- decomposition plan
- issue mappings
- worktree mappings
- agent assignments
- task status
- escalation state
- cleanup state

### GitHub fallback
Important checkpoints must also be written to GitHub:
- research findings
- implementation issue scope
- PR references
- QA findings
- escalation notes
- completion summary

If local hub data is missing, the commander should inspect GitHub issues and PRs to rebuild state.

## 8. Execution Flow

### Phase 1: Intake
1. Human submits task through local interface.
2. Commander creates or updates a local hub record.
3. Commander evaluates complexity, risk, dependencies, and success criteria.

### Phase 2: Research
1. Commander opens one or more research issues.
2. Research agents analyze feasibility, constraints, architecture impact, and risks.
3. Findings are returned in structured issue comments or issue bodies.

### Phase 3: Planning
1. Commander reads research output.
2. Commander decides whether the task should stay single-threaded or be split.
3. Commander creates implementation issues for coder agents.
4. Commander allocates worktrees only when justified.

### Phase 4: Coding
1. Each coder agent receives a scoped issue.
2. Each coder agent works in a dedicated branch and optionally a dedicated worktree.
3. Each coder agent opens or updates a PR tied to the issue.

### Phase 5: QA Loop
1. QA agent reviews the PR against the issue requirements and research findings.
2. QA either:
   - requests changes
   - approves
   - escalates to commander for human verification

### Phase 6: Human Verification
Commander escalates when:
- QA cannot verify correctness
- requirements are ambiguous
- design intent is unclear
- changes are risky or destructive
- infra/security concerns require approval

### Phase 7: Completion and Cleanup
1. Commander confirms QA approval and any required human signoff.
2. Commander posts completion status to GitHub.
3. Commander closes issues.
4. Commander ensures worktree is clean.
5. Commander removes worktree if used.
6. Commander closes branch according to repo policy.

## 9. Parallelization Policy

Commander should prefer parallel coder agents only when all of the following are true:
- task can be decomposed cleanly
- ownership boundaries are clear
- merge conflict risk is low
- available machine resources are sufficient

Commander should prefer single-agent execution when any of the following are true:
- shared files are likely to be edited by multiple agents
- changes are tightly coupled
- disk, CPU, or memory is constrained
- the task is small enough that coordination overhead is wasteful

## 10. Error Handling

### Corrupted or missing hub state
- Reconstruct state from GitHub issues, labels, PRs, and comments.
- Mark the task as reconstructed in the hub record.

### Stalled issue
- Commander marks issue blocked or stale.
- Commander either retries with revised scope or escalates.

### PR review loop failure
- QA returns actionable comments.
- Commander decides whether coder can continue or issue must be re-scoped.

### Human unavailable
- Commander should not pretend completion.
- Commander marks waiting-for-human and preserves current status.

### Dirty worktree at close
- Commander must not delete the worktree before state is explained.
- Commander records uncommitted state and escalates or cleans explicitly.

## 11. Testing Strategy for Future Implementation

The future implementation should be validated with:
- unit tests for task decomposition and risk classification
- unit tests for parallelization decisions
- integration tests for GitHub issue lifecycle
- integration tests for worktree create / clean / remove flow
- recovery tests for corrupted hub reconstruction
- simulation tests for QA escalation and human checkpoints

## 12. Recommended Initial Repository Layout

```text
mothership/
├── docs/
│   └── superpowers/
│       └── specs/
│           └── 2026-04-11-mothership-core-design.md
├── agents/
│   ├── commander.md
│   ├── researcher.md
│   ├── coder.md
│   └── qa.md
├── skills/
│   ├── task-intake-and-decomposition/SKILL.md
│   ├── research-execution/SKILL.md
│   ├── risk-assessment/SKILL.md
│   ├── parallelization-decision/SKILL.md
│   ├── github-issue-management/SKILL.md
│   ├── worktree-management/SKILL.md
│   ├── qa-review-loop/SKILL.md
│   └── human-escalation/SKILL.md
├── registry/
│   └── roles.md
└── hub/
    └── README.md
```

## 13. Recommended v1 Boundary

Build only the orchestration loop for:
- intake
- research
- planning
- coding assignment
- QA review
- human escalation
- cleanup

Avoid adding dashboards, advanced scheduling, or broad provider abstraction in v1.
