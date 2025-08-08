# Usage Guide

This guide explains how to run the CLI, how configuration works, how to define modules, and how to troubleshoot common issues.

## CLI Usage

Build and run:

```bash
# Build
go build -o bin/archwsl-tui-configurator ./cmd/archwsl-tui-configurator

# Run (recommended as root on first run to create user and configure system)
./bin/archwsl-tui-configurator
```

Flags:

- Currently, the CLI does not expose flags. The program launches the TUI directly.

## Configuration

Configuration is persisted to `~/.config/archwsl-tui-configurator/config.yaml` with secure permissions. A saved configuration allows non-interactive re-runs.

Example:

```yaml
Username: dev
GitName: Developer Name
GitEmail: developer@example.com
OhMyZshTheme: agnoster
OhMyZshPlugins: [git, fzf]
DotfilesRepo: "https://example.com/dotfiles.git"
NonInteractive: true
```

Notes:
- `NonInteractive: true` guides the application to apply saved values without prompts
- When `DotfilesRepo` is empty, a default `.zshrc` is written

## Module Definitions (YAML/TOML)

The project includes a strict module loader for YAML and TOML definitions. These APIs validate that every module has a non-empty name and at least one non-empty command.

YAML example:

```yaml
- name: "Go Toolchain"
  description: "Install Go"
  commands:
    - "pacman -S --noconfirm go"
    - "echo done"
- name: "Python"
  commands:
    - "pacman -S --noconfirm python"
```

TOML example:

```toml
[[modules]]
name = "Node"
description = "Install Node"
commands = ["pacman -S --noconfirm nodejs", "npm -v"]

[[modules]]
name = "Dotfiles"
commands = ["echo setup"]
```

## Troubleshooting

- Run as root for first-time provisioning
  - Creating a user, configuring `sudoers`, and firewall rules typically require root

- `ufw` not found or failing
  - Ensure `ufw` is installed: `sudo pacman -S --noconfirm ufw`
  - WSL users: enabling `ufw` may require additional host configuration; the app is idempotent and will re-try safely

- Windows fonts not detected
  - Nerd Font detection reads `/mnt/c/Windows/Fonts`; verify the path exists in your WSL instance
  - Install a Nerd Font on Windows (e.g., JetBrainsMono Nerd Font) and re-run the app

- `pacman` locked or network issues
  - If a lock exists: wait for other package operations to finish or remove stale locks
  - Verify network connectivity and proxy settings

- Git configuration did not apply
  - The app verifies `git config --global user.name`/`user.email`; if verification fails, it reports an error and rolls back in Tx flows

- Idempotency checks
  - Re-running the tool is safe; steps skip when state already matches (e.g., identical `.zshrc`, existing SSH keys, firewall defaults present)

- CI staticcheck version mismatch
  - The repository’s CI pins Go 1.24.6 to avoid toolchain mismatch; ensure your local Go matches `go.mod`

If issues persist, please open an issue with logs and details about your environment.
