# Title: Adopt modular architecture for Arch WSL bootstrap

## Context and Problem Statement
Monolithic scripts are hard to test, extend, and reason about.

## Decision Drivers
- Testability and readability
- Clear separation of concerns
- Safer iteration via CI and linting

## Considered Options
- Keep monolithic scripts
- Split into bash libs + PowerShell module (Chosen)
- Rewrite in another language

## Decision Outcome
Chosen: Split into cohesive bash libraries and a PowerShell module with a single CLI entrypoint.

### Consequences
- Good: Easier testing, reuse, and contributions
- Bad: Initial refactor cost
- Neutral: Slightly more files

### Confirmation
CI passing, feature parity with current behavior, green doctor checks.
