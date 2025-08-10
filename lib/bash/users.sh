# shellcheck shell=bash
# Users and privileges (library)
#
# Description: Create default user, configure passwordless sudo, and ensure
# subordinate id ranges for containers.
# Globals: None

#######################################
# Ensure a user exists with passwordless sudo
# Globals: None
# Args: $1 - username
#######################################
ensure_user_and_sudo() {
  local user="$1"
  if ! id -u "$user" >/dev/null 2>&1; then
    log "*" "Creating user $user with passwordless sudo"
    useradd -m -s /bin/bash "$user"
  fi
  local sudo_file="/etc/sudoers.d/99-${user}-nopasswd"
  echo "${user} ALL=(ALL) NOPASSWD:ALL" > "$sudo_file"
  chmod 0440 "$sudo_file"
}

#######################################
# Ensure subuid/subgid ranges exist for a user
# Globals: None
# Args: $1 - username
#######################################
ensure_subids() {
  local user="$1"
  if ! grep -q "^${user}:" /etc/subuid 2>/dev/null; then
    usermod --add-subuids 524288-589823 "$user" 2>/dev/null \
      || echo "${user}:524288:65536" >> /etc/subuid
  fi
  if ! grep -q "^${user}:" /etc/subgid 2>/dev/null; then
    usermod --add-subgids 524288-589823 "$user" 2>/dev/null \
      || echo "${user}:524288:65536" >> /etc/subgid
  fi
}
