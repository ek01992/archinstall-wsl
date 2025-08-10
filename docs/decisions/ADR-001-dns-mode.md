# ADR-001: DNS mode strategy for Arch WSL

## Status

Accepted

## Context

WSL networking and name resolution can be inconsistent across environments. A reproducible setup needs both stability and flexibility. This repo must operate in network-restricted environments, support Podman networking, and avoid breakage when WSL overwrites `/etc/resolv.conf`.

## Drivers

- Stability across Windows host updates and WSL versions
- WSL constraints (auto-generation of `resolv.conf`)
- Reproducibility for snapshots/exports
- Ease of switching without breaking user config

## Decision

### Support three DNS strategies

- `static` (default): write `/etc/resolv.conf` with configured `NAMESERVERS`, optionally set immutable via `CHATTR_IMMUTABLE_RESOLV=true`
- `resolved`: enable `systemd-resolved`, and provide a boot-time unit to link `/etc/resolv.conf` to the stub
- `wsl`: allow WSL to generate `/etc/resolv.conf` by setting `generateResolvConf=true`

## Consequences

### Pros

- Clear, explicit modes
- `static` offers predictability; `resolved` integrates with systemd; `wsl` preserves default WSL behavior

### Cons

- `static` requires careful switching (must remove immutability before changing)
- `resolved` needs systemd PID1 (thus requires WSL restart after Phase 1)
- `wsl` inherits host/WIN DNS variances

### Switching guidance

- Re-run Phase 1 with `DNS_MODE=<mode>`; Phase 2 will enable services as needed
- For `static` to others, immutability is lifted automatically during the switch

### Confirmation steps

  ```bash
  ls -l /etc/resolv.conf
  readlink -f /etc/resolv.conf 2>/dev/null || true
  command -v resolvectl >/dev/null 2>&1 && resolvectl status || true
  ```
