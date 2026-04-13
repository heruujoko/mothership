# Skill: QA Review Loop

## Purpose

Define the review cycle between coder agents, QA, and commander.

## Flow

1. coder submits PR
2. QA reviews against issue scope and research findings
3. QA either approves, requests changes, or escalates
4. if changes are requested, coder revises and resubmits
5. commander intervenes if scope drift or ambiguity appears

## Rule

The loop continues until one of these is true:
- QA approves
- commander re-scopes the work
- human verification is requested
