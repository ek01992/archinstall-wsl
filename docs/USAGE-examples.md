# Usage examples

## `setup-wsl.ps1` (Windows host)

### Default

```powershell
pwsh -NoProfile -File .\setup-wsl.ps1
```

### Custom distro and user

```powershell
pwsh -NoProfile -File .\setup-wsl.ps1 -DistroName 'arch-dev' -DefaultUser 'alice'
```

### Resolved DNS, more resources

```powershell
$env:WSL_MEMORY = '16GB'
$env:WSL_CPUS   = '8'
pwsh -NoProfile -File .\setup-wsl.ps1 -DnsMode 'resolved' -SwapGB 8
```

### Force reinstall

```powershell
pwsh -NoProfile -File .\setup-wsl.ps1 -Force
```

## `bootstrap.sh` (WSL guest)

### Phase 1 (root)

```bash
DEFAULT_USER=erik DNS_MODE=static OPTIMIZE_MIRRORS=true ./bootstrap.sh phase1
```

### Phase 2 (root, after `wsl --shutdown`)

```bash
DEFAULT_USER=erik DNS_MODE=static PY_VER=3.12.5 ./bootstrap.sh phase2
```

## Dotfiles (user shell)

```bash
bash ./dotfiles/install.sh
```

## Verification snippets

```bash
# PID1/systemd
ps -p 1 -o comm=

# Services
sudo systemctl status podman.socket --no-pager
sudo systemctl status sshd --no-pager

# Toolchains
node -v
python -V
rustc -V
```
