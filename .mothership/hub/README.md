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
- what cleanup remains?

## Format

The hub uses a compact state file plus externalized references for verbose data:

```
.mothership/
  hub/
    state.json          # Compact state machine state (always loaded)
    refs/               # Externalized verbose data (loaded on demand)
      intake.md         # Task summary, risk, success criteria
      research.md       # Research findings (or GitHub issue URL)
      planning.md       # Execution decisions
      coding.md         # Implementation summary (or PR URL)
      qa.md             # QA disposition (or GitHub review URL)
    checkpoints/        # State transition log
      001-intake.json
      002-research.json
      ...
```

### state.json

Compact state machine state (~200-300 tokens). Short keys, references instead of inline content.

```json
{
  "v": 1,
  "task": "T-20260516-001",
  "state": "coding",
  "prev": "planning",
  "entered": "2026-05-16T14:32:00Z",
  "overlay": null,
  "risk": "medium",
  "gh": {
    "issue": "#42",
    "pr": "#43"
  },
  "agents": {
    "coder": "general-purpose"
  },
  "wt": "main",
  "refs": {
    "intake": "refs/intake.md",
    "research": "https://github.com/.../issues/44",
    "planning": "refs/planning.md"
  }
}
```

### Required Fields (hub contract)

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
| `agents` | object | Role → agent_type map for active assignments |
| `wt` | string | Current branch/worktree name |
| `refs` | object | Paths to verbose data in `refs/` or GitHub URLs |

### Overlay Object (only present when active)

```json
{
  "overlay": {
    "type": "blocked",
    "reason": "gh CLI not installed",
    "since": "2026-05-16T15:00:00Z",
    "unblock": "Install gh and authenticate"
  }
}
```

Overlay types:
- `blocked` — requires `type`, `reason`, `since`, `unblock`
- `reconstructed` — requires `type`, `source`, `notes`

### Externalized Data (refs/)

Verbose data lives in `refs/` files or GitHub, referenced by path/URL in state.json:

| Ref File | Content |
|----------|---------|
| `refs/intake.md` | Task summary, risk assessment, success criteria, warmup results, skill preflight |
| `refs/research.md` | Research findings, open questions, risks/constraints |
| `refs/planning.md` | Single/multi-agent decision, decomposition strategy, worktree decision |
| `refs/coding.md` | Implementation notes, files changed, verification performed |
| `refs/qa.md` | QA findings by severity, disposition, required changes |

### Checkpoints

Individual JSON files in `checkpoints/` record each state transition for audit trail and reconstruction:

```json
{
  "seq": 1,
  "from": "intake",
  "to": "research",
  "at": "2026-05-16T14:30:00Z",
  "notes": "Research sub-agent spawned"
}
```

## Reliability Rule

If the hub file is incomplete, inconsistent, or corrupted:
1. commander should inspect GitHub artifacts and `refs/` files
2. commander should reconstruct the missing state from checkpoints
3. commander should mark the hub as reconstructed (overlay)
4. commander should continue only after the reconstructed state is coherent enough

## Persistence Rule

The hub is for live orchestration, not silent reasoning. Important decisions must also be reflected in GitHub artifacts so the workflow remains auditable.

## Startup Hygiene Rule

Commander should execute a warm-up routine during `intake` that:
1. creates runtime artifacts under `.mothership/`
2. enforces `.mothership/` in `.gitignore`
3. records auxiliary skill preflight status for required external skills

## Token Budget

| Component | Tokens |
|-----------|--------|
| state.json (core fields) | ~200-300 |
| Per checkpoint | ~50-100 |
| refs/ files | loaded on demand, not in every context |
| **Total per transition** | **~200-400** (vs ~3000-6500 with inline verbose data) |
