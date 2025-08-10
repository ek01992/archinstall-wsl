# archinstall-wsl

Automated two-phase bootstrap of an Arch Linux WSL distro with reproducible developer toolchains, DNS strategies, and dotfiles.

## Key features

- Two-phase bootstrap:
  - Phase 1: base system, locale/keyring update, packages, WSL config, default user, DNS, dotfiles
  - Phase 2: systemd service enablement and developer toolchains
- DNS modes: `static`, `resolved`, `wsl`
  - `static`: writes `/etc/resolv.conf` with your nameservers, optionally immutable via `chattr`
  - `resolved`: enables `systemd-resolved` and links `resolv.conf` to the stub
  - `wsl`: lets WSL generate `resolv.conf`
- Developer toolchains: `pyenv`, `nvm`, `rustup` (+ `rustfmt`, `clippy`), optional helpers (`neovim`, `direnv`, `starship`), full Podman stack (`podman`, `buildah`, `skopeo`, `slirp4netns`, `netavark`, `fuse-overlayfs`)
- Dotfiles linking from the repo (`bash`, `git`, `nvim`, `rust`, `.editorconfig`)

## Supported environments and prerequisites

- Windows 10 2004+ or Windows 11
- WSL2 enabled
- PowerShell (Windows PowerShell or PowerShell 7)

### Quick start (Windows host)

- End-to-end setup with defaults (`archlinux`, user `erik`, DNS `static`, memory `8GB`, CPUs `4`, swap `4GB`):

  ```powershell
  pwsh -NoProfile -File .\setup-wsl.ps1
  ```

- Advanced usage (custom distro, user, DNS, swap; memory/CPUs via env):

  ```powershell
  $env:WSL_MEMORY = '12GB'
  $env:WSL_CPUS   = '6'
  pwsh -NoProfile -File .\setup-wsl.ps1 -Force `
    -DistroName 'arch-dev' `
    -DefaultUser 'alice' `
    -DnsMode 'resolved' `
    -SwapGB 8
  ```

## What happens

- Phase 1 (run as root inside WSL): locale/keyring refresh, base packages, default user with passwordless sudo, WSL config with `systemd=true`, DNS mode, toolchain bootstraps (installers), and dotfiles linking.
- Phase 2 (run after a WSL restart): enables services (`podman.socket`, `sshd`, and `systemd-resolved` if selected), finalizes toolchains (Node LTS, Python via pyenv, `rustup` components), prints versions summary, and cleans caches.
- Snapshot/export (Windows host):

  ```powershell
  # By default the script also exports the clean snapshot:
  #   $env:USERPROFILE\arch.tar.bak

  # Import later:
  wsl --unregister archlinux
  wsl --import archlinux C:\WSL\Arch $env:USERPROFILE\arch.tar.bak --version 2
  ```

## Configuration

- Environment variables consumed by the bootstrap:
  - `DEFAULT_USER` (default `erik`)
  - `DNS_MODE` (default `static`) one of: `static`, `resolved`, `wsl`
  - `WSL_MEMORY` (default `8GB`) and `WSL_CPUS` (default `4`) used by Windows `.wslconfig` via `setup-wsl.ps1`
  - `REPO_ROOT_MNT` (auto-detected by `setup-wsl.ps1`), used for dotfiles linking
  - `OPTIMIZE_MIRRORS` (default `true`)
  - `CHATTR_IMMUTABLE_RESOLV` (default `true`)
  - `PY_VER` (default `3.12.5` for `pyenv`)
  - `NAMESERVERS` (in `bootstrap.sh` default to `1.1.1.1 9.9.9.9` for static DNS)
- Override options:
  - Windows host: set `WSL_MEMORY`, `WSL_CPUS` before running `setup-wsl.ps1`
  - For other vars, either:
    - Pass them when triggering `bootstrap.sh` manually inside WSL, or
    - Use `setup-wsl.ps1` parameters for `DefaultUser` and `DnsMode` (these are forwarded into Phase 1/2)

## Commands (manual control from inside WSL guest)

- Run Phase 1 manually (as root):

  ```bash
  # Inside WSL (root shell)
  chmod +x ./bootstrap.sh
  DEFAULT_USER=erik DNS_MODE=static OPTIMIZE_MIRRORS=true CHATTR_IMMUTABLE_RESOLV=true ./bootstrap.sh phase1
  ```

- Run Phase 2 manually (after WSL restart):

  ```bash
  # Inside WSL (root shell)
  DEFAULT_USER=erik DNS_MODE=static PY_VER=3.12.5 ./bootstrap.sh phase2
  ```

- Post-install verification (user shell):

  ```bash
  # PID1/systemd
  ps -p 1 -o comm=

  # Services (run via sudo)
  sudo systemctl status podman.socket --no-pager
  sudo systemctl status sshd --no-pager

  # DNS mode specifics
  ls -l /etc/resolv.conf
  readlink -f /etc/resolv.conf 2>/dev/null || true
  command -v resolvectl >/dev/null 2>&1 && resolvectl status || true

  # Toolchains
  node -v || true
  python -V || true
  rustc -V || true

  # Podman quick check (best-effort)
  podman run --rm --network slirp4netns --dns 1.1.1.1 quay.io/podman/hello
  ```

## Outcomes

- Services enabled:
  - Always: `podman.socket`, `sshd`
  - If `DNS_MODE=resolved`: `systemd-resolved` and a `wsl-resolved-link.service` to link `/etc/resolv.conf` to the stub
- Toolchains finalized for the default user:
  - Node LTS via `nvm`, Python `${PY_VER}` via `pyenv` (pip upgraded), Rust stable via `rustup` (+ `rustfmt`, `clippy`)
  - A short versions summary is printed at the end of Phase 2

## Safety and idempotence

- Defensive flags: `set -euo pipefail`
- Retries around pacman operations
- Idempotent helpers: `append_once`, `safe_link`, `ensure_user_and_sudo`, `ensure_subids`
- DNS immutability via `chattr +i` in `static` mode (disabled automatically when switching)

## CI overview

- Linux job:
  - ShellCheck on `lib/bash/*.sh` (and `bin/arch-wsl` if present)
  - `bats` on `tests/bats` if present; otherwise no-op
- Windows job:
  - PSScriptAnalyzer on `src/ps` if present; otherwise skipped

## Folder structure

- `bootstrap.sh`: two-phase orchestrator executed inside WSL
- `setup-wsl.ps1`: Windows host orchestrator (install/unregister, `.wslconfig`, run phases, snapshot)
- `lib/bash/`: modular bash libraries (common helpers, config, system, users, DNS, WSL, services, toolchains, dotfiles)
- `dotfiles/`: opinionated dotfiles linked into the user home (`bash`, `git`, `nvim`, `rust`, `.editorconfig`)
- `.github/workflows/ci.yml`: Shell and PowerShell linting; optional Bats
- Note: `lib/bash/*.sh` provide reusable building blocks. `bootstrap.sh` currently implements its own logic without sourcing these libs.

## FAQ highlights

- Systemd PID 1 in WSL:
  - Phase 1 writes `/etc/wsl.conf` with `systemd=true`; you must restart WSL (`wsl --shutdown`) before Phase 2
- `resolv.conf` behavior:
  - `static`: `/etc/resolv.conf` is written and optionally immutable
  - `resolved`: a unit links it to the resolved stub; service enabled in Phase 2
  - `wsl`: WSL regenerates it; `generateResolvConf=true`
- UNC path caveat:
  - `setup-wsl.ps1` blocks UNC paths (e.g., `\\server\share`) because WSL `/mnt` mapping requires local drive paths
- Pacman keyring hiccups:
  - The script refreshes keyrings and retries; rerun Phase 1 if needed

If anything is ambiguous, prefer environment overrides during manual runs. The optional `config/` mechanism referenced in `lib/bash/config.sh` is not used by `bootstrap.sh` in this repo; use env vars directly.
