# Mothership

Installable package of mothership roles, workflow docs, and skills for use in
other repositories.

## Package Shape

- `agents/`: role contracts for commander, researcher, coder, and QA
- `mothership-config/`: shared role and workflow definitions
- `skills/`: supporting workflow docs plus installable skills for `mothership`,
  `commit`, `commit-push`, and `create-pr`
- `.mothership/hub/`: local hub state contract, checkpoints, and runtime state
- `.mothership/` (runtime, ignored): warm-up artifacts and skill preflight reports

## Manual Entrypoint

Use the `mothership` skill explicitly when you want this orchestration workflow
to run in a mixed-skill environment. This package does not assume mothership is
the only skill installed in a client repository.

On startup, run `skills/mothership/scripts/warmup.sh` to enforce dot-prefixed
runtime artifacts and required auxiliary skill preflight.
