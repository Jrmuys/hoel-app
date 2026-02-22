# MVP Runbook

## Development Principles

- Keep code paths shallow where possible.
- Favor typed domain errors over generic failures.
- Keep comments minimal and only explain non-obvious rationale.

## Operational Checks

- Verify DB migrations apply before starting API routes.
- Treat all third-party calls as unreliable and retry with limits.
- Surface stale integration data in status endpoints.
