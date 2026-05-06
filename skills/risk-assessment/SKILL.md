---
name: risk-assessment
description: Give commander a structured way to classify task risk before choosing execution strategy. Covers risk dimensions and mitigation planning.
---

# Risk Assessment

## Purpose

Give commander a structured way to classify task risk before choosing execution strategy.

## Risk Dimensions

- requirement ambiguity
- blast radius
- shared file overlap
- dependency uncertainty
- infra or security sensitivity
- cleanup complexity

## Risk Levels

### Low
Well-scoped, localized, easy to verify.

### Medium
Some ambiguity or moderate repo impact, but still bounded.

### High
Broad impact, unclear requirements, risky changes, or hard-to-verify outcomes.

## Rule

High-risk tasks should bias toward:
- more research
- fewer parallel agents
- stronger QA scrutiny
- more human checkpoints
