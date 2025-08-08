# ArchWSL TUI Configurator

An idempotent Terminal User Interface (TUI) built in Go using Bubble Tea to provision a clean, repeatable Arch Linux environment on Windows Subsystem for Linux 2 (WSL 2).

This tool turns the manual post-install process into a reliable, test-driven, and aesthetically refined experience. It emphasizes idempotency, transactional safety with rollback, and a modular architecture for toolchains and dotfiles.

## Features

- Idempotent, safe re-runs (steps converge without duplicating work)
- Transactional rollback for critical operations (user, firewall, SSH import, Git, shell)
- TUI built with Bubble Tea with a clear welcome screen and error handling UI (retry/skip/abort)
- System setup building blocks:
  - User creation, add to `wheel`, enable passwordless sudo
  - Git global config (`user.name`, `user.email`) with verification
  - SSH key import from Windows host with correct permissions
  - Firewall (`ufw`) default-deny inbound, allow Windows↔WSL communication
  - Zsh install and Oh My Zsh with theme and plugins
  - Nerd Font detection on Windows host
  - Dotfiles: clone repo, symlink files; fallback to default `.zshrc` if no repo
  - Toolchains (modular): Go, Python (pyenv + pipx), Node.js (nvm + latest LTS)
- Config persistence to `~/.config/archwsl-tui-configurator/config.yaml`; supports non-interactive reruns
- YAML/TOML module loader with strict validation
- CI/CD with lint, vet, staticcheck, tests, and tagged release builds

## Requirements

- Arch Linux on WSL 2 (Windows 11 recommended)
- Go 1.24.6 (pinned)
- Git and internet connectivity

## Environment assumptions and safety notes

- WSL host file access: Operations that read Windows files (e.g., fonts under `/mnt/c/Windows/Fonts`, SSH keys under `/mnt/c/Users/<User>/.ssh`) require that `/mnt/c` is mounted and accessible. The app detects this and falls back safely if unavailable.
- SSH private key import: Copying private keys is sensitive. The library exposes a consent-gated flow; interactive UIs should ask the user before importing. Declining will skip the import.
- `ufw` on WSL: Firewall behavior can vary in WSL. We attempt best-effort configuration and surface errors; consider Windows Defender Firewall rules on the host as an alternative.
- `.wslconfig` tuning: Changes should be done on the Windows host. We recommend a dry-run plan and executing changes via PowerShell with admin rights.

## Install and Build

```bash
# Clone
git clone https://github.com/your-org/archwsl-tui-configurator.git
cd archwsl-tui-configurator

# Build the binary
go build -o bin/archwsl-tui-configurator ./cmd/archwsl-tui-configurator
```

Alternatively, run without building:

```bash
go run ./cmd/archwsl-tui-configurator
```

## Usage

Basic:

```bash
# Run the TUI (as root on a fresh Arch WSL for first-time provisioning)
./bin/archwsl-tui-configurator
```

Configuration persistence:

- The application persists configuration to:
  `~/.config/archwsl-tui-configurator/config.yaml`
- Example config (YAML):

```yaml
Username: dev
GitName: Developer Name
GitEmail: developer@example.com
OhMyZshTheme: agnoster
OhMyZshPlugins: [git, fzf]
DotfilesRepo: "https://example.com/dotfiles.git"
NonInteractive: true
```

When `NonInteractive` is true, the tool can re-apply settings non-interactively using saved values.

Modules (YAML/TOML) definitions are supported at the library level and validated strictly. See `docs/usage.md` for examples.

## Development Workflow

- Specification-driven, test-first development
- Extensive use of seams/mocks to isolate tests from the system
- Idempotency checks for all steps
- Transactional rollback using `internal/tx` (LIFO undo stack) with aggregated rollback error reporting

Quick start for contributors:

```bash
# Install Go 1.24.6 and make sure GOPATH/bin is on PATH
make check-go && make all           # tidy, lint, vet, test (auto-installs golangci-lint if missing)
make build         # build all packages
```

See `CONTRIBUTING.md` for full contributor guidelines.

## Testing and CI/CD

Local:

```bash
go vet ./...
$(go env GOPATH)/bin/staticcheck ./...
go test -v ./...
```

CI:

- `.github/workflows/ci.yml` runs on push/PR:
  - Go 1.24.6
  - Lint (`golangci-lint`), vet, `staticcheck`, tests (race, coverage), and coverage threshold
- `.github/workflows/release.yml` builds on tag push `v*`:
  - Produces `linux/amd64` and `linux/arm64` binaries and creates a GitHub Release

Releasing:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

## Documentation

- Architecture: see `docs/architecture.md`
- Usage and examples: see `docs/usage.md`
- Contributing: see `CONTRIBUTING.md`

## License

No license file is currently provided. Add a license of your choice (e.g., MIT, Apache-2.0) as `LICENSE` at the repository root.
