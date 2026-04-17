# Roles Registry

This file is the single source of truth for what roles exist in mothership v1, what each role is for, and when the commander may invoke them.

New roles should be added here before they are introduced into the workflow.

## Core Rules

- Only the commander communicates directly with the human operator.
- All non-commander roles report through structured local hub updates and GitHub artifacts.
- Every role must have a clear input contract, output contract, and escalation rule.
- Roles may be extended later, but v1 assumes a fixed core set.

## Available Roles

### Commander
**Purpose:** Central orchestrator and sole human-facing role.  
**Invoked when:** A new task enters the system or an existing task needs replanning, escalation handling, or completion.  
**Primary outputs:** hub state updates, GitHub issue creation, assignment decisions, escalation notices, completion summary.  
**Key constraints:** Must not lose task state silently; must prefer clear audit trail over hidden reasoning.

### Research Agent
**Purpose:** Validate requirements, gather context, identify risks, and provide decision support before implementation planning.  
**Invoked when:** Commander needs feasibility checks, architecture context, prior art, dependency review, or ambiguity analysis.  
**Primary outputs:** research issue updates, structured findings, recommendations, unanswered questions.  
**Key constraints:** Should not implement code or make workflow decisions reserved for commander.

### Coder Agent
**Purpose:** Implement a scoped issue in a branch and optional worktree.  
**Invoked when:** Commander has defined implementation work and assigned an issue.  
**Primary outputs:** code changes, commit history, PR, implementation notes.  
**Key constraints:** Must stay within issue scope and report blockers rather than silently broadening scope.

### QA Agent
**Purpose:** Review implementation against issue requirements and research findings.  
**Invoked when:** A coder agent has produced a PR or significant implementation checkpoint.  
**Primary outputs:** approval, requested changes, escalation recommendation, review summary.  
**Key constraints:** Must not approve ambiguous or unverifiable changes; must escalate clearly when human judgment is required.

### PR Monkey
**Purpose:** Maintain pull requests through the GitHub CLI, including PR creation, feedback triage, reply posting, and thread resolution after commander-routed fixes land.  
**Invoked when:** Commander receives a PR to maintain, a request to create a PR from the current project, or a request to process PR review feedback.  
**Primary outputs:** PR URL and description, structured feedback items for commander, reply and resolution updates, push confirmations, blockers related to `gh` or PR state.  
**Key constraints:** Must use `gh`, must work feedback one item at a time, must route actionable feedback back through commander, and must never merge without explicit human consent.

## Future Candidate Roles

### Designer
Review UX, flows, wireframes, and product interaction details.

### Security Reviewer
Audit security-sensitive logic, secrets handling, auth flows, and risk exposure.

### Documentation Writer
Produce or improve developer docs, release notes, onboarding, and operator guidance.

### DevOps Agent
Handle CI/CD, infrastructure wiring, deployment pipelines, and operational tooling.

## Role Introduction Process

To add a new role:
1. Add the role to this registry.
2. Add a matching agent file in `agents/`.
3. Define any supporting skill docs in `skills/`.
4. Update commander workflow rules if the role changes orchestration behavior.
