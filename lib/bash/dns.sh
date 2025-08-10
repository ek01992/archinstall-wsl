# DNS configuration (library)
#
# Description: Manage resolv.conf in static, resolved, or WSL-managed modes.
# Globals: NAMESERVERS_A, CHATTR_IMMUTABLE_RESOLV

#######################################
# Configure static /etc/resolv.conf
# Globals: NAMESERVERS_A, CHATTR_IMMUTABLE_RESOLV
# Args: None
#######################################
configure_dns_static() {
  log "*" "Configuring static resolv.conf"
  command -v chattr >/dev/null 2>&1 && chattr -i /etc/resolv.conf 2>/dev/null || true
  rm -f /etc/resolv.conf || true
  {
    local ns
    for ns in "${NAMESERVERS_A[@]}"; do
      printf 'nameserver %s\n' "$ns"
    done
    echo "options edns0"
  } > /etc/resolv.conf
  if [[ "${CHATTR_IMMUTABLE_RESOLV}" == "true" ]] && command -v chattr >/dev/null 2>&1; then
    chattr +i /etc/resolv.conf || true
  fi
}

#######################################
# Configure systemd-resolved
# Globals: None
# Args: None
#######################################
configure_dns_resolved() {
  log "*" "Configuring systemd-resolved support"
  command -v chattr >/dev/null 2>&1 && chattr -i /etc/resolv.conf 2>/dev/null || true
  rm -f /etc/resolv.conf || true

  cat <<'EOF' > /etc/systemd/system/wsl-resolved-link.service
[Unit]
Description=WSL: Link /etc/resolv.conf to systemd-resolved stub
After=systemd-resolved.service
ConditionPathIsSymbolicLink=!/etc/resolv.conf

[Service]
Type=oneshot
ExecStart=/usr/bin/ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload || true
}

#######################################
# Use WSL-managed resolv.conf
# Globals: None
# Args: None
#######################################
configure_dns_wsl() {
  log "*" "Using WSL-managed resolv.conf"
  command -v chattr >/dev/null 2>&1 && chattr -i /etc/resolv.conf 2>/dev/null || true
  rm -f /etc/resolv.conf || true
}

#######################################
# Public switch for DNS configuration
# Globals: None
# Args: $1 - dns mode
#######################################
configure_dns() {
  case "${1:-static}" in
    static)   configure_dns_static ;;
    resolved) configure_dns_resolved ;;
    wsl)      configure_dns_wsl ;;
    *)        log "i" "Unknown DNS_MODE='${1}', defaulting to static"; configure_dns_static ;;
  esac
}

#######################################
# Post-configuration summary for resolved mode
# Globals: None
# Args: $1 - dns mode
#######################################
post_dns_summary_if_resolved() {
  local mode="${1:-}"
  if [[ "${mode}" == "resolved" ]]; then
    command -v resolvectl >/dev/null 2>&1 && log "i" "systemd-resolved active. Check: resolvectl status"
    [[ -L /etc/resolv.conf ]] && log "i" "/etc/resolv.conf -> $(readlink -f /etc/resolv.conf)"
  fi
}
