# Architecture

## Windows host orchestrator: `setup-wsl.ps1`

- Ensures WSL2 availability and sets default version to 2
- Writes user `.wslconfig` with memory/CPU/swap
- Prepares paths for WSL `/mnt` mapping; guards against UNC paths
- Installs the distro, runs bootstrap phases, and exports a clean snapshot

## Linux guest bootstrap: `bootstrap.sh` (two phases)

- Phase 1: Locale/keyring, base packages, default user, WSL config, DNS mode, toolchain installers, dotfiles
- Phase 2: Enable services, finalize toolchains, cleanup, and print versions

## Modular Bash libraries in `lib/bash/`

- `common.sh`: logging, retry, `append_once`, sudo fallback, guards
- `config.sh`: config file loading and boolean normalization (not used by `bootstrap.sh` at present)
- `system.sh`: locale, mirrors, pacman update, package install
- `users.sh`: create user, passwordless sudo, subuid/subgid ranges
- `dns.sh`: static/resolved/wsl strategies and linking unit
- `wsl.sh`: `wsl.conf` generation, systemd check
- `services.sh`: enable services, cleanup, quick doctor checks
- `toolchains.sh`: install/finalize `pyenv`, `nvm`, `rustup`
- `dotfiles.sh`: safe symlinks, per-user linking

> Note: `bootstrap.sh` currently implements its own logic and does not source these libs; the libs enable extension and reuse.

## Data and control flows

### Variables from Windows host to Linux guest

- `setup-wsl.ps1` sets/forwards:
  - `.wslconfig`: `WSL_MEMORY`, `WSL_CPUS`, `SwapGB`
  - Phase 1/2 env: `DEFAULT_USER`, `DNS_MODE`, also forwards repo mount path (`REPO_ROOT_MNT`)

### DNS mode branching

- `static`: write `/etc/resolv.conf` with `NAMESERVERS`, optional `chattr +i`
- `resolved`: install linking unit `wsl-resolved-link.service`; enable `systemd-resolved` in Phase 2
- `wsl`: remove `/etc/resolv.conf` and set WSL to regenerate

### Systemd dependency

- Phase 1 writes `/etc/wsl.conf` with `systemd=true`; you must `wsl --shutdown` before Phase 2 so PID 1 is `systemd`

## Idempotency and retry

- Idempotent helpers:
  - `append_once` for config lines
  - `safe_link` and backups with `.bak`
  - `ensure_user_and_sudo` and `ensure_subids`
- Retries:
  - Pacman keyring refresh and updates are wrapped in `retry`
- Defensive shell flags:
  - `set -euo pipefail`, best-effort error guards (`|| true`) where safe

## Security considerations

- Passwordless sudo:
  - A scoped file `/etc/sudoers.d/99-<user>-nopasswd` is created for the default user
- DNS immutability:
  - In `static` mode, `/etc/resolv.conf` can be made immutable via `chattr +i`
- No SUID/SUDO in scripts:
  - Scripts do not set SUID bits; privileged operations occur in root context during bootstrap

## Extensibility

- Add packages:
  - Extend the package arrays in `bootstrap.sh` or, if wiring to libs later, via config/env processed by `lib/bash/config.sh`
- Add services/hooks:
  - Extend `enable_services` in `bootstrap.sh` or `lib/bash/services.sh`
- Dotfiles layering:
  - Add files under `dotfiles/`, which are symlinked into the user home; backups are created when needed.
