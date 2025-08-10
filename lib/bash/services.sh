# shellcheck shell=bash
# Services and cleanup (library)
#
# Description: Enable services under systemd and perform cleanup for snapshot.
# Globals: None

#######################################
# Enable required services (systemd must be PID1)
# Globals: None
# Args: $1 - dns mode
#######################################
enable_services() {
  local dns_mode="${1:-}"
  log "*" "Enabling services (requires systemd)"
  if ! is_systemd_active; then
    die "systemd is not active (PID 1 != systemd). Restart WSL then run phase2."
  fi

  echo 'ListenAddress 127.0.0.1' > /etc/ssh/sshd_config.d/loopback.conf
  systemctl daemon-reload || true

  if [[ "${dns_mode}" == "resolved" ]]; then
    systemctl enable --now systemd-resolved || true
    systemctl enable --now wsl-resolved-link.service || true
  fi

  systemctl enable --now podman.socket
  systemctl enable --now sshd
}

#######################################
# Clean caches and journal to minimize snapshot size
# Globals: None
# Args: None
#######################################
cleanup_for_snapshot() {
  log "*" "Cleaning caches and logs"
  yes | pacman -Scc --noconfirm >/dev/null 2>&1 || true
  rm -rf "$HOME/.cache"/* 2>/dev/null || true
  rm -rf /var/cache/pacman/pkg/* 2>/dev/null || true
  journalctl --rotate >/dev/null 2>&1 || true
  journalctl --vacuum-time=1s >/dev/null 2>&1 || true
}

#######################################
# Simple environment checks
# Globals: None
# Args: $1 - user, $2 - dns_mode
#######################################
doctor_checks() {
  local user="$1" dns_mode="$2"
  log "*" "Doctor: quick checks"
  if in_wsl; then
    log "+" "WSL detected"
  else
    err "WSL not detected"
  fi
  if is_systemd_active; then
    log "+" "systemd running (PID1)"
  else
    err "systemd not PID1"
  fi
  if command -v pacman >/dev/null 2>&1; then
    log "+" "pacman present"
  else
    err "pacman missing"
  fi
  if [[ -f /etc/wsl.conf ]]; then
    log "+" "wsl.conf present"
  else
    err "wsl.conf missing"
  fi
  case "${dns_mode}" in
    static)
      if [[ -f /etc/resolv.conf ]]; then
        log "+" "resolv.conf exists"
      else
        err "resolv.conf missing"
      fi
      ;;
    resolved)
      if systemctl is-enabled systemd-resolved >/dev/null 2>&1; then
        log "+" "systemd-resolved enabled"
      else
        err "resolved not enabled"
      fi
      ;;
    wsl) log "i" "WSL will manage resolv.conf" ;;
  esac
  if id -u "${user}" >/dev/null 2>&1; then
    log "+" "user ${user} exists"
  else
    err "user ${user} missing"
  fi
}
