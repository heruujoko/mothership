---
name: superpowers
description: Bundled dependency skill used by mothership research for broad context gathering and option mapping.
---

# obra/superpowers

## Intended Usage During Research

When invoked by a researcher sub-agent during mothership's research phase, use
this before `obra/making-plans` packages the handoff:

1. **Context expansion** — Read relevant files, docs, recent commits. Map the
   current state of the codebase related to the task. Understand existing
   patterns, architecture, and conventions before proposing changes.

2. **Option mapping** — Identify 2-3 viable approaches. For each:
   - Describe what it does and how it works
   - List trade-offs (complexity, risk, maintenance, performance)
   - Note dependencies and prerequisites
   - Give your recommendation with reasoning

3. **Assumption validation** — List assumptions implicit in the task request.
   For each assumption:
   - Mark as **confirmed** if codebase evidence supports it
   - Mark as **unverified** if no evidence exists
   - Mark as **contradicted** if codebase evidence refutes it

4. **Scope assessment** — If the task spans multiple independent subsystems,
   recommend decomposition. Identify which pieces are independent, how they
   relate, and what order they should be built.

## Output Format

Write findings in the research-execution template:

- **summary** — What you investigated and what you found
- **assumptions** — Confirmed, unverified, and contradicted assumptions
- **constraints** — Known technical or business constraints
- **risks** — Risks and their likelihood/impact
- **recommendation** — Your recommended approach with alternatives
- **open questions** — Ambiguities for commander to escalate to human

## Boundaries

- Do NOT write the final user-approved design spec. That belongs to the
  brainstorming and planning flow.
- Do NOT ask iterative clarifying questions. Work with the research issue
  context; note ambiguities as open questions.
- Do NOT implement code or make workflow decisions reserved for commander.

## Coordination

During mothership research phase, invoke this skill after
`obra/brainstorming` and before `obra/making-plans`. The research-execution
skill defines the output format; this skill provides the process for gathering
context and comparing approaches.

## Note

In environments with an upstream obra package, this local marker may be replaced
or shadowed by the upstream installation.
