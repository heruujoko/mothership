# Skill: Human Escalation

## Purpose

Define when and how commander requests human input.

## Trigger Conditions

- requirement ambiguity
- QA cannot verify correctness
- risky or destructive change
- security or infrastructure impact
- unresolved product or design judgment
- cleanup anomaly requiring explicit approval

## Expected Commander Output

A concise escalation should include:
- what happened
- why automation cannot proceed safely
- what decision is needed from the human
- current issue / PR / worktree state
- recommended options if available

## Rule

Escalation is not failure. It is the correct action when confidence is insufficient.
