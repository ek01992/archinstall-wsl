# Architecture

This document describes the high-level architecture, major components, and how the specification maps to the test suite. The application is designed for idempotency, transactional rollback, and clear separation of concerns.

## System Overview

Entry point and UI:

- `cmd/archwsl-tui-configurator/main.go`: program entry; instantiates `internal/app` and runs the Bubble Tea program
- `internal/app`: minimal app wrapper that starts the TUI
- `internal/ui`: Bubble Tea model with a descriptive welcome view and an error handling UI (retry / skip / abort)

Core capabilities (each implemented as DI Services with small, testable interfaces):

- `internal/user`: Service for user creation, password, wheel group, sudoers; Zsh and Oh My Zsh installation; Tx variants
- `internal/git`: Service for global `user.name` and `user.email` with verification; Tx variant
- `internal/ssh`: Service for importing Windows SSH keys with correct permissions; idempotent copy
- `internal/firewall`: Service to configure `ufw` (default-deny inbound, allow Windows↔WSL subnet); idempotent; Tx variant
- `internal/toolchain`:
  - `golang`: Service to install/update Go toolchain; verify `go version`
  - `python`: Service to install `pyenv`, configure Python version, ensure `pipx`
  - `nodejs`: Service to install `nvm`, latest LTS Node.js as default
- `internal/nerdfont`: Service to detect presence of a Nerd Font on Windows host (via FS/Runner/Platform DI)
- `internal/dotfiles`: clone and link dotfiles, or write a default `.zshrc`
- `internal/modules`: YAML/TOML module loader with normalization and validation
- `internal/config`: save/load persisted config (`~/.config/archwsl-tui-configurator/config.yaml`) with strict permissions
- `internal/platform`: Service for environment detection (`IsWSL`, `CanEditHostFiles`, `IsMounted`, `Getenv`)
- `internal/pacman`: helpers to query local packages
- `internal/tx`: generic transaction manager for LIFO rollback

Cross-cutting concerns:

- Idempotency checks before state changes
- Dependency Injection (DI) with interfaces (`Runner`, `FS`, `Env`, `Platform`, etc.) and adapters between packages
- Transactional wrappers (Tx) provide compensating actions for rollback on failure
- Logging via `internal/logx` at critical boundaries

## DI Wiring (Provider)

Production wiring is centralized in `internal/app/provider.go`:

- Constructs `runtime` dependencies: `Runner`, `FS`, `Env`
- Builds `platform.Service` with runtime FS/Env adapters
- Adapts `runtime` abstractions into each service’s interfaces (e.g., nodejs `Runner` with `Shell`)
- Exposes a `Provider` struct containing ready-to-use services (`User`, `SSH`, `Git`, `Firewall`, `NerdFont`, `GoToolchain`, `NodeToolchain`, `PythonToolchain`, `Platform`)

This enables orchestration layers (e.g., UI flows) to depend on services, not on global seams.

## Component Relationships (Diagram Description)

1. The CLI entrypoint initializes `internal/app`, which starts the Bubble Tea `internal/ui` model
2. The UI (or orchestrator) receives a `Provider` and calls service methods (instead of package-level functions)
3. Each service depends on interfaces (FS/Runner/Env/Platform), injected via `Provider`
4. Config is persisted and retrieved via `internal/config`
5. Optional module definitions are parsed via `internal/modules` (YAML/TOML) for validated, normalized command lists

Visually, imagine layers:

- UI layer: `internal/ui` and `internal/app`
- Orchestration layer: constructs and uses `Provider`
- Capability layer: `internal/*` DI services providing actions + Tx wrappers
- Foundation layer: `internal/runtime`, `internal/tx`, `internal/pacman`

## Spec-to-Test Mapping

The project is specification-driven. Each requirement is implemented alongside tests that prove correctness, idempotency, and concurrency safety. The table below summarizes the mapping.

| **Spec Ref** | **Requirement**                             | **Test Cases**                                                                                                  |
| ------------ | ------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| A.1          | Aesthetically pleasing Bubble Tea interface | TUI renders without error; all views fit terminal window; theme colors match spec palette                       |
| A.2          | Interactive configuration flow              | User can navigate forward/back; selecting defaults applies expected values; manual entry persists               |
| A.3          | State detection                             | Detect installed packages; detect default shell; detect existing users; skip steps if already satisfied         |
| B.1          | User account creation                       | Creates user with correct username; user in `wheel` group; passwordless sudo enabled                            |
| B.2          | Git configuration                           | `git config --global` returns expected `user.name` and `user.email`                                             |
| B.3          | SSH key import                              | Keys copied from Windows host; permissions correct (`700` dir, `600` private keys); fails gracefully if no keys |
| B.4          | Zsh + Oh My Zsh                             | Zsh installed; set as default shell; Oh My Zsh installed with correct theme/plugins                             |
| B.5          | Nerd Font detection                         | Detects installed Nerd Font; prompts if missing; skips prompt if present                                        |
| C.1          | Module loader                               | Loads YAML/TOML module definition; detects invalid file; hot-load works without rebuild                         |
| C.2          | Go toolchain setup                          | Installs latest Go; sets GOPATH; verifies `go version`                                                          |
| C.3          | Python toolchain setup                      | Installs pyenv; installs Python version; installs pipx; verifies installation                                   |
| C.4          | Node.js toolchain setup                     | Installs nvm; installs latest LTS; verifies node/npm version                                                    |
| C.5          | Dotfile management                          | Clones repo; symlinks files; default dotfiles applied when no repo provided                                     |
| D.1          | `.wslconfig` optimization                   | Writes/updates file on Windows host; CPU/memory settings match spec                                             |
| D.2          | Arch update & essentials                    | Runs `pacman -Syu`; installs essential packages; handles already-installed case                                 |
| D.3          | Firewall config                             | `ufw` enabled; default-deny inbound; allows Windows↔WSL                                                         |
| D.4          | Maintenance scripts                         | Scripts copied; marked executable; verified in PATH                                                             |
| E.1          | Config persistence                          | `config.yaml` created; rerun in non-interactive mode uses saved config                                          |
| E.2          | Rollback mechanism                          | On step failure, rolls back changes; restores pre-step state                                                    |
| E.3          | Failure handling UI                         | Presents retry/skip/abort options; logs error clearly                                                           |
| F.1          | Concurrency safety                          | `Service_Concurrent_NoRaces` tests for user/git/ssh/firewall; `-race` CI job is clean                           |
| IV.1         | Test-first                                  | Unit/integration tests exist for each spec item before implementation                                           |
| IV.2         | CI/CD pipeline                              | Linting passes; static analysis passes; all tests pass; coverage threshold enforced                             |

Note: Historical seam-based helpers were removed in favor of DI; new code must use services and injected dependencies.
