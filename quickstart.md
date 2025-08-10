# Quick start (Windows host)

- Default install:

  ```powershell
  pwsh -NoProfile -File .\setup-wsl.ps1
  ```

- Choose DNS mode:

  ```powershell
  pwsh -NoProfile -File .\setup-wsl.ps1 -DnsMode static     # or resolved | wsl
  ```

- Rerun Phase 2 (if you tweak config and want to re-finalize):

  ```powershell
  # Phase 2 is run automatically by setup-wsl.ps1.
  # To re-run:
  wsl -d archlinux -u root -- bash -lc "cd /mnt/$(($pwd.Path.Substring(0,1)).ToLower())$($pwd.Path.Substring(2) -replace '\\','/') ; DEFAULT_USER='erik' DNS_MODE='static' ./bootstrap.sh phase2"
  ```

## Common variants

- Custom distro name, user, CPUs/memory/swap:

  ```powershell
  $env:WSL_MEMORY = '12GB'
  $env:WSL_CPUS   = '6'
  pwsh -NoProfile -File .\setup-wsl.ps1 -Force `
    -DistroName 'arch-dev' `
    -DefaultUser 'alice' `
    -DnsMode 'resolved' `
    -SwapGB 8
  ```

- Run `bootstrap.sh` directly inside WSL:

  ```bash
  # Phase 1 (root)
  chmod +x ./bootstrap.sh
  DEFAULT_USER=erik DNS_MODE=static ./bootstrap.sh phase1

  # Restart WSL from Windows:
  #   wsl --shutdown

  # Phase 2 (root)
  DEFAULT_USER=erik DNS_MODE=static PY_VER=3.12.5 ./bootstrap.sh phase2
  ```

## Export/import snapshot (Windows host)

```powershell
# Export a clean snapshot
wsl --terminate archlinux
wsl --export archlinux $env:USERPROFILE\arch.tar.bak

# Import later
wsl --unregister archlinux
wsl --import archlinux C:\WSL\Arch $env:USERPROFILE\arch.tar.bak --version 2
```

## Post-install sanity checks (WSL guest)

```bash
# systemd
ps -p 1 -o comm=

# DNS
ls -l /etc/resolv.conf
readlink -f /etc/resolv.conf 2>/dev/null || true
command -v resolvectl >/dev/null 2>&1 && resolvectl status || true

# Services
sudo systemctl status podman.socket --no-pager
sudo systemctl status sshd --no-pager

# Toolchains
node -v || true
python -V || true
rustc -V || true

# Podman smoke test
podman run --rm --network slirp4netns --dns 1.1.1.1 quay.io/podman/hello
```
