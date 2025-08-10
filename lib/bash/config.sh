# Configuration loader for Arch WSL (library)
#
# Description: Loads config from config/defaults.env and config/local.env,
# normalizes booleans, and converts space-delimited variables into arrays.
# Globals set: DEFAULT_USER, WSL_MEMORY, WSL_CPUS, DNS_MODE,
#              CHATTR_IMMUTABLE_RESOLV, OPTIMIZE_MIRRORS,
#              BASE_PACKAGES_A, CONTAINER_PACKAGES_A,
#              OPTIONAL_LANG_PACKAGES_A, OPTIONAL_DOTFILES_HELPERS_A,
#              NAMESERVERS_A

#######################################
# Normalize a truthy/falsy string to true/false
# Args: $1 - input string
# Outputs: normalized value to STDOUT
#######################################
normalize_bool() {
  local v="${1:-}"
  case "${v,,}" in
    1|true|yes|y|on) echo "true" ;;
    0|false|no|n|off|"") echo "false" ;;
    *) echo "${v}" ;;
  esac
}

#######################################
# Split a string into words (bash) and print one per line
# Args: $1 - input string
# Outputs: items to STDOUT (one per line)
#######################################
split_words_to_array() {
  local s="${1:-}"
  # shellcheck disable=SC2206
  local arr=( $s )
  printf '%s\n' "${arr[@]}"
}

#######################################
# Load configuration files and materialize arrays
# Globals set: see file header
# Args: None
#######################################
load_config() {
  local cfg_dir="${ROOT_DIR:-.}/config"
  [[ -f "${cfg_dir}/defaults.env" ]] && . "${cfg_dir}/defaults.env"
  [[ -f "${cfg_dir}/local.env" ]] && . "${cfg_dir}/local.env"

  DNS_MODE="${DNS_MODE:-static}"
  CHATTR_IMMUTABLE_RESOLV="$(normalize_bool "${CHATTR_IMMUTABLE_RESOLV:-true}")"
  OPTIMIZE_MIRRORS="$(normalize_bool "${OPTIMIZE_MIRRORS:-true}")"

  mapfile -t BASE_PACKAGES_A < <(split_words_to_array "${BASE_PACKAGES:-}")
  mapfile -t CONTAINER_PACKAGES_A < <(split_words_to_array "${CONTAINER_PACKAGES:-}")
  mapfile -t OPTIONAL_LANG_PACKAGES_A < <(split_words_to_array "${OPTIONAL_LANG_PACKAGES:-}")
  mapfile -t OPTIONAL_DOTFILES_HELPERS_A < <(split_words_to_array "${OPTIONAL_DOTFILES_HELPERS:-}")
  mapfile -t NAMESERVERS_A < <(split_words_to_array "${NAMESERVERS:-}")
}
