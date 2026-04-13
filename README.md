# Mothership

Installable package of mothership roles, workflow docs, and skills for use in
other repositories.

## Package Shape

- `agents/`: role contracts for commander, researcher, coder, and QA
- `mothership-config/`: shared role and workflow definitions
- `skills/`: supporting workflow docs plus the installable `mothership` skill entrypoint
- `hub/`: local hub state contract

## Manual Entrypoint

Use the `mothership` skill explicitly when you want this orchestration workflow
to run in a mixed-skill environment. This package does not assume mothership is
the only skill installed in a client repository.
