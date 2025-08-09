# Contributing to ArchWSL TUI Configurator

Thanks for your interest in improving this project! This document explains how to set up a local environment, the branching and commit conventions, how to write tests, and how to submit a pull request.

## Local Development Setup

Prerequisites:

- Arch Linux or compatible Linux environment (WSL 2 recommended for parity)
- Go 1.24.6 (pinned)
- Git

Clone and build:

```bash
git clone https://github.com/your-org/archwsl-tui-configurator.git
cd archwsl-tui-configurator
make check-go && make all     # tidy, lint, vet, test (auto-installs golangci-lint if missing)
```

Useful commands:

```bash
make tidy          # go mod tidy
make lint          # golangci-lint run ./...
make vet           # go vet ./...
make test          # go test -v ./...
make build         # go build ./...
```

Notes:

- Tests use fakes/mocks to avoid touching real system state. You do NOT need root to run tests.
- Ensure `$(go env GOPATH)/bin` is on your PATH if you call `staticcheck` directly.
- Go version is pinned in CI and docs for reproducibility. Use `make check-go` to verify your local version.

## Design & Architecture Guidelines (DI)

- Do not introduce new package-level global seams (e.g., `var runCommand = func(...)`).
- Add functionality to DI Services and inject runtime dependencies via interfaces defined in `internal/runtime` and service packages.
- For orchestration, construct services via `internal/app.Provider`, which wires production `Runner`, `FS`, and `Env`.
- Prefer small, focused interfaces and adapters to bridge between packages.

## Testing Guidelines

- Test-first: write unit tests before implementation.
- Maintain strict separation between tests and implementation; do not alter tests to create false positives.
- Use provided fake implementations in service tests rather than overriding globals.
- Exercise idempotency and rollback (Tx) paths; add concurrency tests where relevant (e.g., `Service_Concurrent_NoRaces`).

## Branching Strategy

- Use short-lived feature branches off `main` (GitHub Flow)
  - Example: `feat/git-config-rollback`, `fix/ssh-idempotency`
- Keep PRs focused and atomic (aligned to a single testable requirement)
- Rebase on `main` when necessary; avoid merge commits in PRs

## Commit Message Conventions

Follow Conventional Commits:

- `feat`: new user-visible feature
- `fix`: bug fix
- `docs`: documentation updates
- `test`: add or refactor tests
- `refactor`: code change that neither fixes a bug nor adds a feature
- `chore`: maintenance tasks (deps, tooling), no production code change
- `ci`/`build`: CI/CD or build system changes

Examples:

```text
feat(user): add transactional rollback to user.Service
fix(ssh): ensure idempotent copy skips identical files
docs: document DI wiring via app.Provider
```

## Writing and Running Tests

Principles:

- Test-first: write unit tests before implementation
- Maintain strict separation between tests and implementation; do not alter tests to create false positives
- Prefer DI with fakes over global seams; do not introduce new global seams
- Ensure idempotency and rollback behaviors are exercised; include concurrency tests where appropriate

Run tests locally:

```bash
go test -v ./...
```

Static analysis:

```bash
go vet ./...
$(go env GOPATH)/bin/staticcheck ./...
```

## Pull Request Process

1. Ensure `make all` passes (tidy, lint, vet, tests; coverage and race checks in CI)
2. Update documentation (`README.md`, `docs/*`) if behavior changes
3. Ensure changes are atomic and well-tested
4. Open a PR against `main` with a clear title and description
5. Request review; address feedback promptly
6. CI must be green before merge

Thanks again for contributing!
