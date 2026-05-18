---
name: research-execution
description: Define how research agents investigate a task before implementation planning. Covers hypothesis formation, context gathering, and structured output.
---

# Research Execution

## Purpose

Define how research agents investigate a task before implementation planning.

## Responsibilities

- validate assumptions
- gather architecture context
- identify risks and constraints
- compare implementation options
- return a concise recommendation

## Skill Coordination

During mothership research phase, invoke `obra/brainstorming` in research mode
to explore the task, `obra/superpowers` for context expansion and option
mapping, and `obra/making-plans` to turn findings into a planning-ready
handoff before writing findings. This skill defines the output format.

## Output Template

- summary
- assumptions
- constraints
- risks
- recommendation
- open questions
