# Local Hub State

The local hub is the operational memory of mothership. It is the live communication center for commander and spawned agents.

GitHub remains the durable external record for important checkpoints.

## Purpose

The hub should make it possible for commander to answer:
- what task is active?
- which agents are assigned?
- what issues and PRs exist?
- what is blocked?
- what is waiting on human input?
- what resources are allocated?
- what cleanup remains?

## Recommended Fields

- task_id
- parent_task_summary
- status
- created_at
- updated_at
- risk_level
- decomposition_strategy
- parallelization_decision
- resource_snapshot
- issues
- pull_requests
- agent_assignments
- worktrees
- escalations
- cleanup_status
- reconstruction_flag
- warmup_checklist
- skill_preflight

## Reliability Rule

If the hub file is incomplete, inconsistent, or corrupted:
1. commander should inspect GitHub artifacts
2. commander should reconstruct the missing state
3. commander should mark the hub as reconstructed
4. commander should continue only after the reconstructed state is coherent enough

## Persistence Rule

The hub is for live orchestration, not silent reasoning. Important decisions must also be reflected in GitHub artifacts so the workflow remains auditable.

## Startup Hygiene Rule

Commander should execute a warm-up routine during `intake` that:
1. creates runtime artifacts under `.mothership/`
2. enforces `.mothership/` in `.gitignore`
3. records auxiliary skill preflight status for required external skills

## Suggested Future Format

A future implementation may store the hub as JSON, YAML, or a small directory of structured files. v1 design does not force the exact format.
