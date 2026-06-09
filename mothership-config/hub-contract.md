# Hub Contract

This file is the canonical, version-controlled copy of the hub contract.
It is mirrored at `.mothership/hub/README.md` at runtime.

**Do not edit `.mothership/hub/README.md` directly** — edit this file instead,
and copy it to the runtime location during warmup.

---

## Design

**Single file, no external refs.** All state — core fields, phase data, agent
status, decisions — lives in one YAML file. There are no separate `refs/` files,
checkpoint logs, or format duplicates. One read loads everything the commander
needs. One write saves everything.

```
.mothership/hub/
  state.yaml    # Single source of truth — all state, all phase data
  README.md     # This file (mirrored from mothership-config/hub-contract.md)
```

### state.yaml

```yaml
v: 1
task: T-20260516-001
state: coding
prev: planning
entered: 2026-05-16T14:32:00Z
overlay: null
risk: medium
gh:
  issue: "#42"
  pr: "#43"
agents:
  coder: active
wt: main
summary: Short task description
wiki:
  stale_pages: []

# --- phase data (inlined per transition) ---
intake:
  success: |
    - Criteria items
    - One per line
  warmup: preflight completed
research:
  summary: Key findings in one line
  constraints:
    - Pi-only
    - Claude/Codex unchanged
  recommendations:
    - Update registries
planning:
  execution: single-agent
  rationale: Tightly coupled scope
  worktree: none
  scope:
    - Itemized plan steps
coding:
  implemented: |
    - What was done
    - Files changed summary
  verified: verification methods used
  branch: feature-branch
```

### Required Fields

| Key | Type | Description |
|-----|------|-------------|
| `v` | int | Format version for future compatibility |
| `task` | string | Task ID |
| `state` | string | Current workflow state |
| `prev` | string | Previous workflow state |
| `entered` | string | ISO timestamp when state was entered |
| `overlay` | object\|null | Active overlay (blocked/reconstructed) or null |

### Recommended Fields

| Key | Type | Description |
|-----|------|-------------|
| `risk` | enum | `low`, `medium`, `high`, `critical` |
| `gh` | object | GitHub references: `issue`, `pr` |
| `agents` | object | Role → status/agent_type map |
| `wt` | string | Current branch/worktree name |
| `summary` | string | One-line task summary |
| `wiki` | object | Top-level wiki runtime data such as `stale_pages` |
| `intake` | object | Success criteria, warmup, risk notes |
| `research` | object | Findings, constraints, recommendations |
| `planning` | object | Execution shape, scope, rationale |
| `coding` | object | Implementation summary, files, verification |

### Overlay Object (only present when active)

```yaml
overlay:
  type: blocked
  reason: gh CLI not installed
  since: 2026-05-16T15:00:00Z
  unblock: Install gh and authenticate
```

Overlay types:
- `blocked` — requires `type`, `reason`, `since`, `unblock`
- `reconstructed` — requires `type`, `source`, `notes`

## State Transitions

Each transition updates `state`, `prev`, `entered`, and appends or replaces the
phase section for the new state. Prior phase sections are preserved for context.

History is implicit: `state` → `prev` gives the current direction. For deeper
history, inspect GitHub artifacts (issues, PRs, commits).

## Reliability Rule

If `state.yaml` is missing, incomplete, or corrupted:
1. Commander should inspect GitHub artifacts
2. Commander should reconstruct the missing state
3. Commander should mark the hub as reconstructed (overlay)
4. Commander should continue only after the reconstructed state is coherent enough

## Persistence Rule

The hub is for live orchestration, not silent reasoning. Important decisions
must also be reflected in GitHub artifacts so the workflow remains auditable.

## Token Budget

| Component | Tokens |
|-----------|--------|
| state.yaml (core fields) | ~200-300 |
| state.yaml (all phase data) | ~400-800 |
| **Total per session** | **~400-800** (was ~3000-6500 with refs) |
