# System setup for Arch WSL (library)
#
# Description: Locale configuration, pacman keyring refresh, mirror tuning,
# and base package installation.
# Globals: OPTIMIZE_MIRRORS, BASE_PACKAGES_A, CONTAINER_PACKAGES_A,
#          OPTIONAL_LANG_PACKAGES_A, OPTIONAL_DOTFILES_HELPERS_A

#######################################
# Ensure en_US.UTF-8 locale is generated and exported
# Globals: None
# Args: None
#######################################
ensure_locale() {
  log "*" "Configuring locale (en_US.UTF-8)"
  sed -i 's/^# *en_US.UTF-8/en_US.UTF-8/' /etc/locale.gen
  locale-gen
  echo 'LANG=en_US.UTF-8' | tee /etc/locale.conf >/dev/null
  append_once 'LANG=en_US.UTF-8' /etc/environment
  export LANG=en_US.UTF-8
  append_once 'export LANG=en_US.UTF-8' /root/.bashrc
}

#######################################
# Optionally optimize mirrors (reflector) and set ParallelDownloads
# Globals: OPTIMIZE_MIRRORS
# Args: None
#######################################
optimize_mirrors_if_enabled() {
  if [[ "${OPTIMIZE_MIRRORS}" != "true" ]]; then return 0; fi

  if ! grep -q '^\[options\]' /etc/pacman.conf; then
    printf "\n[options]\n" | tee -a /etc/pacman.conf >/dev/null
  fi

  awk '
    BEGIN{inopt=0; inserted=0}
    /^\[/{ inopt = ($0 ~ /^\[options\]/); print; next }
    {
      if (inopt && $1 ~ /^ParallelDownloads/) next
      print
      if (inopt && !inserted) { print "ParallelDownloads = 10"; inserted=1 }
    }' /etc/pacman.conf > /etc/pacman.conf.tmp
  mv /etc/pacman.conf.tmp /etc/pacman.conf

  log "*" "Installing reflector to optimize mirrors (best-effort)"
  pacman -Sy --noconfirm --noprogressbar || true
  pacman -S --needed --noconfirm --noprogressbar rsync reflector || true

  if command -v reflector >/dev/null 2>&1; then
    if ! reflector --latest 20 --sort rate --save /etc/pacman.d/mirrorlist; then
      log "i" "reflector failed; keeping existing mirrorlist"
    else
      log "*" "Mirrorlist updated by reflector"
    fi
  fi
}

#######################################
# Refresh keyring and update system
# Globals: None
# Args: None
#######################################
pacman_quiet_update() {
  log "*" "Refreshing keyring and updating base system"
  command -v timedatectl >/dev/null 2>&1 && timedatectl status >/dev/null 2>&1 || true
  rm -rf /etc/pacman.d/gnupg || true
  pacman-key --init || true
  pacman-key --populate archlinux || true
  retry 3 3 pacman -Sy --noconfirm --noprogressbar archlinux-keyring || true
  pacman-key --populate archlinux || true
  optimize_mirrors_if_enabled
  retry 3 5 pacman -Syyu --noconfirm --noprogressbar
}

#######################################
# Install core and optional packages
# Globals: BASE_PACKAGES_A, CONTAINER_PACKAGES_A,
#          OPTIONAL_LANG_PACKAGES_A, OPTIONAL_DOTFILES_HELPERS_A
# Args: None
#######################################
install_packages() {
  log "*" "Installing packages (combined transactions)"
  local pkgs=("${BASE_PACKAGES_A[@]}" "${CONTAINER_PACKAGES_A[@]}")
  retry 3 5 pacman -S --needed --noconfirm --noprogressbar "${pkgs[@]}"
  log "i" "Optional language toolchains"
  pacman -S --needed --noconfirm --noprogressbar "${OPTIONAL_LANG_PACKAGES_A[@]}" || true
  log "i" "Optional dotfiles helpers (best-effort)"
  pacman -S --needed --noconfirm --noprogressbar "${OPTIONAL_DOTFILES_HELPERS_A[@]}" || true
}
