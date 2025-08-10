# Troubleshooting

## systemd is not PID 1 after Phase 1

- Symptom: Phase 2 fails with a message about systemd not active.
- Fix:

  ```powershell
  # Windows host
  wsl --shutdown
  ```

  Then re-run Phase 2:

  ```bash
  # Inside WSL (root)
  ./bootstrap.sh phase2
  ```

  Validate:

  ```bash
  ps -p 1 -o comm=
  ```

## DNS failures

### Switching modes

  ```bash
  # Inside WSL (root)
  DNS_MODE=static   ./bootstrap.sh phase1   # rewrites resolv.conf (may set immutable)
  DNS_MODE=resolved ./bootstrap.sh phase1   # sets up linking unit; Phase 2 enables resolved
  DNS_MODE=wsl      ./bootstrap.sh phase1   # removes resolv.conf; WSL will regenerate
  ```

### Check resolv.conf

  ```bash
  ls -l /etc/resolv.conf
  readlink -f /etc/resolv.conf 2>/dev/null || true
  ```

### Resolved diagnostics

  ```bash
  sudo systemctl status systemd-resolved --no-pager
  resolvectl status
  ```

### pyenv/nvm/rustup issues

- Re-run finalization safely:

  ```bash
  DEFAULT_USER=erik PY_VER=3.12.5 ./bootstrap.sh phase2
  ```

- Check shell init lines exist in `~/.bashrc` and `~/.profile`. If needed, re-link dotfiles:

  ```bash
  bash ./dotfiles/install.sh
  ```

### Podman networking under WSL

- Smoke test:

  ```bash
  podman system migrate || true
  podman run --rm --network slirp4netns --dns 1.1.1.1 quay.io/podman/hello
  ```

- If DNS issues persist, try `DNS_MODE=static` and re-run Phase 1/2.

### UNC path caveat

- If the repo or temp path is a UNC path (e.g., `\\server\share`), `setup-wsl.ps1` will abort.
- Move the repo to a local drive (e.g., `C:\Users\you\projects\...`) and re-run.

### Pacman keyring failures

- The script retries keyring refresh and update. If still failing, re-run Phase 1:

  ```bash
  ./bootstrap.sh phase1
  ```

### Diagnostics and verification

- Services:

  ```bash
  sudo systemctl status sshd --no-pager
  sudo systemctl status podman.socket --no-pager
  ```

- DNS symlink:

  ```bash
  ls -l /etc/resolv.conf
  readlink -f /etc/resolv.conf 2>/dev/null || true
  ```

- Journal and logs:

  ```bash
  sudo journalctl -u sshd -b --no-pager | tail -n 100
  sudo journalctl -u systemd-resolved -b --no-pager | tail -n 100
  ```
