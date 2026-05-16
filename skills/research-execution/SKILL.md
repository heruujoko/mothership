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

During mothership research phase, invoke `obra/superpowers` for context
expansion and option mapping before writing findings. The superpowers skill
provides the process for gathering context and comparing approaches; this
skill defines the output format.

## Output Template

- summary
- assumptions
- constraints
- risks
- recommendation
- open questions
