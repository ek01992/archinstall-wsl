# Arch WSL Bootstrap (Modular)

## Overview

This repository bootstraps an Arch Linux WSL distro using a modular bash CLI and a PowerShell module. It installs core packages, configures WSL and DNS, creates a user, links dotfiles, and sets up developer toolchains.

## Quick start

 From Windows PowerShell (as normal user):

```powershell
# Optional: set overrides in env vars
$env:WSL_MEMORY = '8GB'; $env:WSL_CPUS = '4'

# Run the orchestrator (will use the module under src/ps)
./setup-wsl.ps1 -DistroName archlinux -DefaultUser erik -DnsMode static -SwapGB 4 -Force
```

## This script will

- Write `.wslconfig`
- Install or reset the distro
- Run Phase 1 and Phase 2 inside WSL using the `bin/arch-wsl` CLI
- Export a snapshot to `%USERPROFILE%\arch.tar.bak`

## Configuration

- Defaults live in `config/defaults.env`.
- Create `config/local.env` (copy from `config/local.env.example`) to override locally.
- Environment variables also override at runtime.

## Commands (inside WSL)

```bash
sudo bin/arch-wsl phase1
sudo bin/arch-wsl phase2
sudo bin/arch-wsl doctor
```

## Development

- Run ShellCheck and bats locally or rely on the GitHub Actions CI.
- Tests live under `tests/bats` and should remain separate from implementation per standards.

## Decisions

- See `docs/adr/0001-modularization.md` for the modularization ADR.
