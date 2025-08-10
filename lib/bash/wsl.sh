# shellcheck shell=bash
# WSL helpers (library)
#
# Description: WSL configuration and PID1/systemd checks.
# Globals: None

#######################################
# Check if systemd is PID 1
# Globals: None
# Returns: 0 if systemd is PID1, non-zero otherwise
#######################################
is_systemd_active() {
  local pid1
  pid1="$(ps -p 1 -o comm= 2>/dev/null || true)"
  [[ "${pid1}" == "systemd" ]]
}

#######################################
# Write /etc/wsl.conf with requested defaults
# Globals: None
# Args: $1 - default user; $2 - dns mode (static|resolved|wsl)
#######################################
configure_wsl() {
  local user="$1" dns_mode="$2"
  log "*" "Writing /etc/wsl.conf (systemd=true, resolv: ${dns_mode}, default user)"
  local gen="true"
  case "${dns_mode}" in
    static|resolved) gen="false" ;;
    wsl) gen="true" ;;
    *) log "i" "Unknown DNS_MODE='${dns_mode}', defaulting to static"; gen="false"; dns_mode="static" ;;
  esac

  cat <<EOF > /etc/wsl.conf
[boot]
systemd=true

[user]
default=${user}

[network]
generateResolvConf=${gen}
EOF
}
