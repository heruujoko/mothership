---
name: qa-review-loop
description: Define the review cycle between coder agents, QA, and commander. Covers feedback routing, iteration, and approval criteria.
---

# QA Review Loop

## Purpose

Define the review cycle between coder agents, QA, and commander.

## Flow

1. coder submits PR
2. QA reviews against issue scope and research findings
3. QA either approves, requests changes, or escalates
4. if changes are requested, coder revises and resubmits
5. if feedback originates from GitHub review threads, PR monkey relays one
   actionable item at a time back to commander
6. commander intervenes if scope drift or ambiguity appears

## Rule

The loop continues until one of these is true:
- QA approves
- commander re-scopes the work
- human verification is requested

## PR Feedback Rule

External PR feedback does not bypass the QA loop.

PR monkey may read, post, and resolve GitHub review threads, but the underlying
requested change must still be routed through commander and the normal
research/coding/QA discipline.
