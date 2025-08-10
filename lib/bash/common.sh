# Common helpers for Arch WSL scripts (library)
#
# Description: Logging, error handling, retries, and small utilities shared
# across bash modules. No shebang; must be sourced.
# Globals: none

set -o pipefail

#######################################
# Timestamp in YYYY-MM-DD HH:MM:SS
# Globals: None
# Outputs: timestamp to STDOUT
#######################################
TS() { date +'%F %T'; }

#######################################
# Structured log to STDOUT
# Globals: None
# Args: $1 - level symbol (*, +, !, i); $2 - message
#######################################
log() { printf "[%s] %s %s\n" "$(TS)" "${1}" "${2}"; }

#######################################
# Error log to STDERR
# Globals: None
# Args: $@ - message
#######################################
err() { printf "[%s] ! %s\n" "$(TS)" "$*" >&2; }

#######################################
# Exit with error after logging
# Globals: None
# Args: $@ - message
#######################################
die() { err "$*"; exit 1; }

# Trap errors with location context
trap 'err "Failure at line $LINENO in ${FUNCNAME[0]:-main}"' ERR

#######################################
# Retry a command N times with delay
# Globals: None
# Args: $1 - tries, $2 - delay_seconds, $3.. - command
# Returns: 0 on success; 1 if exhausted
#######################################
retry() {
  local tries="$1" delay="$2"; shift 2
  local n=1
  until "$@"; do
    if (( n >= tries )); then return 1; fi
    sleep "${delay}"
    n=$((n+1))
  done
}

#######################################
# Append a line to a file only once
# Globals: None
# Args: $1 - line, $2 - file
#######################################
append_once() {
  local line="$1" file="$2"
  mkdir -p "$(dirname "$file")"
  if [[ ! -f "$file" ]]; then
    printf '%s\n' "$line" >"$file"
    return 0
  fi
  if ! grep -qxF -- "$line" "$file" 2>/dev/null; then
    printf '%s\n' "$line" >>"$file"
  fi
}

#######################################
# Ensure a sudo fallback exists when sudo is not installed
# Globals: None
# Args: None
#######################################
ensure_sudo_fallback() {
  if ! command -v sudo >/dev/null 2>&1; then
    sudo() {
      if [[ "$#" -ge 3 && "$1" == "-u" ]]; then
        shift
        local __sudo_user="$1"; shift
        su -s /bin/bash - "$__sudo_user" -c "$*"
      else
        "$@"
      fi
    }
  fi
}

#######################################
# Require execution as root (EUID 0)
# Globals: None
# Args: None
#######################################
require_root() { (( EUID == 0 )) || die "This command must be run as root."; }

#######################################
# Detect if running in WSL
# Globals: None
# Returns: 0 if yes, 1 otherwise
#######################################
in_wsl() { grep -qi microsoft /proc/version 2>/dev/null; }

#######################################
# Load version from VERSION file or emit dev version
# Globals: ROOT_DIR
# Outputs: version to STDOUT
#######################################
load_version() {
  if [[ -f "${ROOT_DIR:-.}/VERSION" ]]; then
    cat "${ROOT_DIR}/VERSION"
  else
    echo "0.0.0-dev"
  fi
}
