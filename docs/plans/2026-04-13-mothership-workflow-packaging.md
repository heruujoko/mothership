# Mothership Workflow Packaging Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a single editable commander workflow registry and a manual `mothership` skill entrypoint for mixed-skill environments.

**Architecture:** Keep orchestration flow out of role prose by defining states and transitions in a dedicated YAML registry file. Add a top-level installable `mothership` skill that explicitly activates this package, while keeping the existing role and workflow docs as internal building blocks.

**Tech Stack:** Markdown, YAML

---

### Task 1: Add the commander workflow registry

**Files:**
- Create: `mothership-config/workflow.yaml`

**Step 1: Define the canonical commander states**
- Add `intake`, `research`, `planning`, `coding`, `qa`, `waiting_for_human`, `complete`, and `cleanup`.

**Step 2: Define allowed transitions and exit criteria**
- Record the intended phase order and the main artifacts required before leaving each state.

**Step 3: Add non-phase overlays**
- Add `blocked` and `reconstructed` as overlays rather than normal workflow phases.

### Task 2: Update commander documentation

**Files:**
- Modify: `agents/commander.md`

**Step 1: Point commander to the workflow registry**
- State that the workflow order lives in `mothership-config/workflow.yaml`.

**Step 2: Add explicit state handling language**
- Describe normal phases and overlay handling for `blocked` and `reconstructed`.

### Task 3: Add the manual mothership entrypoint skill

**Files:**
- Create: `skills/mothership/SKILL.md`

**Step 1: Define the trigger**
- Make it clear this skill is for explicit invocation such as `/mothership` in a mixed-skill environment.

**Step 2: Define what the skill loads**
- Reference the role registry, commander workflow, hub contract, and supporting workflow docs in this package.

**Step 3: Define the activation rule**
- State that this skill is the opt-in front door and should not assume Mothership is the only skill installed.

### Task 4: Add package-level guidance

**Files:**
- Modify: `README.md`

**Step 1: Describe the package**
- Explain that this repository is an installable bundle of roles and skills.

**Step 2: Document the entrypoint**
- Note that `mothership` is the explicit manual skill for activating the orchestration workflow in another repo.
