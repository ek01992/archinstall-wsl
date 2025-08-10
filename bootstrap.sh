#!/usr/bin/env bash
set -euo pipefail

# ============================================
# Arch WSL two-phase bootstrap
#  - Phase 1: Locale, keyring, base packages, WSL config, user creation, dev tools init, dotfiles
#  - Phase 2: Enable services under systemd, finalize user toolchains, cleanup
# ============================================

PHASE="${1:-phase1}"
DEFAULT_USER="${DEFAULT_USER:-erik}"
WSL_MEMORY="${WSL_MEMORY:-8GB}"
WSL_CPUS="${WSL_CPUS:-4}"
REPO_ROOT_MNT="${REPO_ROOT_MNT:-}"

# DNS strategy: static | resolved | wsl
DNS_MODE="${DNS_MODE:-static}"
# DNS settings when we manage resolv.conf ourselves (static mode)
NAMESERVERS=("1.1.1.1" "9.9.9.9")
CHATTR_IMMUTABLE_RESOLV="${CHATTR_IMMUTABLE_RESOLV:-true}"

# Package selections
BASE_PACKAGES=(
  base-devel curl git wget unzip zip jq ripgrep fzf tmux htop lsof rsync openssh net-tools
)
CONTAINER_PACKAGES=(
  crun conmon containers-common aardvark-dns netavark slirp4netns fuse-overlayfs podman buildah skopeo
)
OPTIONAL_LANG_PACKAGES=( go )
OPTIONAL_DOTFILES_HELPERS=( neovim direnv starship )
OPTIMIZE_MIRRORS="${OPTIMIZE_MIRRORS:-true}"

# --------------------------------------------
# Helpers
# --------------------------------------------
TS() { date +'%F %T'; }
log() { printf "[%s] %s %s\n" "$(TS)" "${1}" "${2}"; }
fail() { log "!" "${1}"; exit 1; }

retry() {
  local tries="$1" delay="$2"; shift 2
  local n=1
  until "$@"; do
    if (( n >= tries )); then
      return 1
    fi
    sleep "$delay"
    n=$((n+1))
  done
}

# Provide sudo fallback when not installed (Phase 1 runs as root).
# Support "sudo -u ..." and plain "sudo ..."
if ! command -v sudo >/dev/null 2>&1; then
  sudo() {
    if [ "$#" -ge 3 ] && [ "$1" = "-u" ]; then
      shift
      local __sudo_user="$1"; shift
      su -s /bin/bash - "$__sudo_user" -c "$*"
    else
      "$@"
    fi
  }
fi

append_once() {
  local line="$1" file="$2"
  mkdir -p "$(dirname "$file")"
  if [ ! -f "$file" ]; then
    printf '%s\n' "$line" >"$file"
    return 0
  fi
  if ! grep -qxF -- "$line" "$file" 2>/dev/null; then
    printf '%s\n' "$line" >>"$file"
  fi
}

safe_link() {
  # link with backup if dest exists and is not the same link
  local src="$1" dest="$2"
  mkdir -p "$(dirname "$dest")"
  if [ -e "$dest" ] && [ ! -L "$dest" ]; then
    mv -n "$dest" "${dest}.bak" || true
  fi
  ln -sfn "$src" "$dest"
}

safe_link_user() {
  local user="$1" src="$2" dest="$3"
  sudo -u "$user" bash -lc "mkdir -p \"\$(dirname \"$dest\")\""
  # If a regular file exists, back it up
  if sudo -u "$user" test -e "$dest" && ! sudo -u "$user" test -L "$dest"; then
    sudo -u "$user" mv -n "$dest" "${dest}.bak" || true
  fi
  sudo -u "$user" ln -sfn "$src" "$dest"
}

# --------------------------------------------
# System configuration and packages
# --------------------------------------------
ensure_locale() {
  log "*" "Configuring locale (en_US.UTF-8)"
  sudo sed -i 's/^# *en_US.UTF-8/en_US.UTF-8/' /etc/locale.gen
  sudo locale-gen
  echo 'LANG=en_US.UTF-8' | sudo tee /etc/locale.conf >/dev/null
  append_once 'LANG=en_US.UTF-8' /etc/environment
  export LANG=en_US.UTF-8
  append_once 'export LANG=en_US.UTF-8' /root/.bashrc
}

optimize_mirrors_if_enabled() {
  if [ "${OPTIMIZE_MIRRORS}" != "true" ]; then
    return 0
  fi

  # Ensure [options] exists
  if ! grep -q '^\[options\]' /etc/pacman.conf; then
    printf "\n[options]\n" | sudo tee -a /etc/pacman.conf >/dev/null
  fi

  # Normalize ParallelDownloads under [options]
  sudo awk '
    BEGIN{inopt=0; inserted=0}
    /^\[/{ inopt = ($0 ~ /^\[options\]/); print; next }
    {
      if (inopt && $1 ~ /^ParallelDownloads/) next
      print
      if (inopt && !inserted) { print "ParallelDownloads = 10"; inserted=1 }
    }
  ' /etc/pacman.conf | sudo tee /etc/pacman.conf.tmp >/dev/null
  sudo mv /etc/pacman.conf.tmp /etc/pacman.conf

  log "*" "Installing reflector to optimize mirrors (best-effort)"
  sudo pacman -Sy --noconfirm --noprogressbar || true
  sudo pacman -S --needed --noconfirm --noprogressbar rsync reflector || true

  if command -v reflector >/dev/null 2>&1; then
    if ! reflector --latest 20 --sort rate --save /etc/pacman.d/mirrorlist; then
      log "i" "reflector failed; keeping existing mirrorlist"
    else
      log "*" "Mirrorlist updated by reflector"
    fi
  fi
}

pacman_quiet_update() {
  log "*" "Refreshing keyring and updating base system"
  if command -v timedatectl >/dev/null 2>&1; then
    timedatectl status >/dev/null 2>&1 || true
  fi
  sudo rm -rf /etc/pacman.d/gnupg || true
  sudo pacman-key --init || true
  sudo pacman-key --populate archlinux || true
  retry 3 3 sudo pacman -Sy --noconfirm --noprogressbar archlinux-keyring || true
  sudo pacman-key --populate archlinux || true
  optimize_mirrors_if_enabled
  retry 3 5 sudo pacman -Syyu --noconfirm --noprogressbar
}

install_packages() {
  log "*" "Installing packages (combined transactions)"
  local pkgs=("${BASE_PACKAGES[@]}" "${CONTAINER_PACKAGES[@]}")
  retry 3 5 sudo pacman -S --needed --noconfirm --noprogressbar "${pkgs[@]}"
  log "i" "Optional language toolchains"
  sudo pacman -S --needed --noconfirm --noprogressbar "${OPTIONAL_LANG_PACKAGES[@]}" || true
  log "i" "Optional dotfiles helpers (best-effort)"
  sudo pacman -S --needed --noconfirm --noprogressbar "${OPTIONAL_DOTFILES_HELPERS[@]}" || true
}

ensure_user_and_sudo() {
  local user="$1"
  if ! id -u "$user" >/dev/null 2>&1; then
    log "*" "Creating user $user with passwordless sudo"
    sudo useradd -m -s /bin/bash "$user"
  fi
  local sudo_file="/etc/sudoers.d/99-${user}-nopasswd"
  echo "${user} ALL=(ALL) NOPASSWD:ALL" | sudo tee "$sudo_file" >/dev/null
  sudo chmod 0440 "$sudo_file"
}

ensure_subids() {
  local user="$1"
  if ! grep -q "^${user}:" /etc/subuid 2>/dev/null; then
    sudo usermod --add-subuids 524288-589823 "$user" 2>/dev/null \
      || echo "${user}:524288:65536" | sudo tee -a /etc/subuid >/dev/null
  fi
  if ! grep -q "^${user}:" /etc/subgid 2>/dev/null; then
    sudo usermod --add-subgids 524288-589823 "$user" 2>/dev/null \
      || echo "${user}:524288:65536" | sudo tee -a /etc/subgid >/dev/null
  fi
}

configure_wsl() {
  log "*" "Writing /etc/wsl.conf (systemd=true, resolv.conf mode: ${DNS_MODE}, default user)"
  local gen="true"
  case "${DNS_MODE}" in
    static|resolved) gen="false" ;;
    wsl) gen="true" ;;
    *) log "i" "Unknown DNS_MODE='${DNS_MODE}', defaulting to static"; gen="false"; DNS_MODE="static" ;;
  esac

  cat <<EOF | sudo tee /etc/wsl.conf >/dev/null
[boot]
systemd=true

[user]
default=${DEFAULT_USER}

[network]
generateResolvConf=${gen}
EOF
}

configure_dns_static() {
  log "*" "Configuring static resolv.conf"
  if command -v chattr >/dev/null 2>&1; then
    sudo chattr -i /etc/resolv.conf 2>/dev/null || true
  fi
  sudo rm -f /etc/resolv.conf || true
  {
    for ns in "${NAMESERVERS[@]}"; do
      printf 'nameserver %s\n' "$ns"
    done
    echo "options edns0"
  } | sudo tee /etc/resolv.conf >/dev/null
  if [ "${CHATTR_IMMUTABLE_RESOLV}" = "true" ] && command -v chattr >/dev/null 2>&1; then
    sudo chattr +i /etc/resolv.conf || true
  fi
}

configure_dns_resolved() {
  log "*" "Configuring systemd-resolved support"
  if command -v chattr >/dev/null 2>&1; then
    sudo chattr -i /etc/resolv.conf 2>/dev/null || true
  fi
  sudo rm -f /etc/resolv.conf || true

  # Service that links resolv.conf to resolved stub at boot
  cat <<'EOF' | sudo tee /etc/systemd/system/wsl-resolved-link.service >/dev/null
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
  sudo systemctl daemon-reload || true
}

configure_dns_wsl() {
  log "*" "Using WSL-managed resolv.conf"
  if command -v chattr >/dev/null 2>&1; then
    sudo chattr -i /etc/resolv.conf 2>/dev/null || true
  fi
  sudo rm -f /etc/resolv.conf || true
  # Let WSL regenerate on next start
}

configure_dns() {
  case "${DNS_MODE}" in
    static)   configure_dns_static ;;
    resolved) configure_dns_resolved ;;
    wsl)      configure_dns_wsl ;;
    *)        log "i" "Unknown DNS_MODE='${DNS_MODE}', defaulting to static"; DNS_MODE="static"; configure_dns_static ;;
  esac
}

# --------------------------------------------
# User toolchains and dotfiles
# --------------------------------------------
install_pyenv_nvm_rustup_for_user() {
  local user="$1"
  local home_dir
  home_dir="$(eval echo ~"$user")"

  log "*" "Setting up pyenv, nvm, rustup for ${user}"

  # Ensure shell rc files exist and are owned by the user BEFORE installers amend them
  sudo -u "$user" bash -lc 'touch "$HOME/.bashrc" "$HOME/.profile"'
  sudo chown "$user:$user" "${home_dir}/.bashrc" "${home_dir}/.profile" 2>/dev/null || true

  # pyenv
  if [ ! -d "${home_dir}/.pyenv" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc 'set -e; command -v curl >/dev/null 2>&1; curl -fsSL https://pyenv.run | bash'
  fi

  # Ensure shells load pyenv
  append_once 'export PYENV_ROOT="$HOME/.pyenv"' "${home_dir}/.bashrc"
  append_once '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' "${home_dir}/.bashrc"
  append_once 'eval "$(pyenv init -)"' "${home_dir}/.bashrc"
  append_once 'eval "$(pyenv virtualenv-init -)"' "${home_dir}/.bashrc"
  append_once 'export PYENV_ROOT="$HOME/.pyenv"' "${home_dir}/.profile"
  append_once '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' "${home_dir}/.profile"
  append_once 'eval "$(pyenv init -)"' "${home_dir}/.profile"

  # nvm
  if [ ! -d "${home_dir}/.nvm" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc 'set -e; command -v curl >/dev/null 2>&1; curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash'
  fi
  append_once 'export NVM_DIR="$HOME/.nvm"' "${home_dir}/.bashrc"
  append_once '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' "${home_dir}/.bashrc"
  append_once '[ -s "$NVM_DIR/bash_completion" ] && . "$NVM_DIR/bash_completion"' "${home_dir}/.bashrc"
  append_once 'export NVM_DIR="$HOME/.nvm"' "${home_dir}/.profile"
  append_once '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' "${home_dir}/.profile"

  # rustup
  if [ ! -x "${home_dir}/.cargo/bin/rustc" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc 'set -e; command -v curl >/dev/null 2>&1; curl --proto "=https" --tlsv1.2 -fsSL https://sh.rustup.rs | sh -s -- -y'
  fi
  append_once 'export PATH="$HOME/.cargo/bin:$PATH"' "${home_dir}/.bashrc"

  # Ensure ownership for future edits
  sudo chown "$user:$user" "${home_dir}/.bashrc" "${home_dir}/.profile" 2>/dev/null || true
  sudo chown -R "$user:$user" "${home_dir}/.pyenv" "${home_dir}/.nvm" "${home_dir}/.cargo" 2>/dev/null || true
}

link_dotfiles_for_user() {
  local user="$1"
  local home_dir
  home_dir="$(eval echo ~"$user")"
  local dotroot=""

  if [ -n "$REPO_ROOT_MNT" ] && [ -d "$REPO_ROOT_MNT/dotfiles" ]; then
    dotroot="${REPO_ROOT_MNT}/dotfiles"
  else
    if command -v powershell.exe >/dev/null 2>&1; then
      local win_cwd
      win_cwd="$(powershell.exe -NoProfile -Command '$pwd.Path' | tr -d '\r')"
      if [ -n "$win_cwd" ]; then
        local repo_root
        repo_root="$(wslpath -u "$win_cwd")"
        [ -d "${repo_root}/dotfiles" ] && dotroot="${repo_root}/dotfiles"
      fi
    fi
  fi

  if [ -n "$dotroot" ] && [ -d "$dotroot" ]; then
    log "*" "Linking dotfiles for ${user} from ${dotroot}"
    sudo -u "$user" bash -lc "mkdir -p \"$HOME/.config\" \"$HOME/.cargo\" \"$HOME/.config/nvim\""
    safe_link_user "$user" "$dotroot/bash/.bashrc" "$home_dir/.bashrc"
    safe_link_user "$user" "$dotroot/bash/.bash_aliases" "$home_dir/.bash_aliases"
    safe_link_user "$user" "$dotroot/git/.gitconfig" "$home_dir/.gitconfig"
    safe_link_user "$user" "$dotroot/nvim/init.vim" "$home_dir/.config/nvim/init.vim"
    safe_link_user "$user" "$dotroot/rust/config.toml" "$home_dir/.cargo/config.toml"
    safe_link_user "$user" "$dotroot/editorconfig/.editorconfig" "$home_dir/.editorconfig"
    log "*" "Dotfiles linked."
  else
    log "i" "dotfiles folder not found; skipped"
  fi
}

# --------------------------------------------
# Services and finalize
# --------------------------------------------
is_systemd_active() {
  local pid1
  pid1="$(ps -p 1 -o comm= 2>/dev/null || true)"
  [ "$pid1" = "systemd" ]
}

enable_services() {
  log "*" "Enabling services (requires systemd)"
  if ! is_systemd_active; then
    fail "systemd is not active (PID 1 != systemd). Restart WSL with systemd=true and rerun phase2."
  fi

  echo 'ListenAddress 127.0.0.1' | sudo tee /etc/ssh/sshd_config.d/loopback.conf >/dev/null
  sudo systemctl daemon-reload || true

  if [ "${DNS_MODE}" = "resolved" ]; then
    sudo systemctl enable --now systemd-resolved || true
    sudo systemctl enable --now wsl-resolved-link.service || true
  fi

  sudo systemctl enable --now podman.socket
  sudo systemctl enable --now sshd
}

finalize_user_toolchains() {
  local user="$1"
  local PY_VER="${PY_VER:-3.12.5}"

  log "*" "Finalizing toolchains for ${user} (Node LTS, Python ${PY_VER}, pip upgrade, rust components)"

  # nvm + Node LTS
  sudo -u "$user" bash -lc '
    set -e
    [ -f "$HOME/.profile" ] && . "$HOME/.profile"
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
    nvm install --lts
    nvm alias default lts/*
  '

  # pyenv + Python
  sudo -u "$user" bash -lc "
    set -e
    [ -f \"\$HOME/.profile\" ] && . \"\$HOME/.profile\"
    [ -f \"\$HOME/.bashrc\" ] && . \"\$HOME/.bashrc\"
    export PYENV_ROOT=\"\$HOME/.pyenv\"
    [ -d \"\$PYENV_ROOT/bin\" ] && export PATH=\"\$PYENV_ROOT/bin:\$PATH\"
    if command -v pyenv >/dev/null 2>&1; then
      eval \"\$(pyenv init -)\"
      pyenv install -s ${PY_VER} || { echo 'Warning: pyenv failed to install ${PY_VER}, will use system python'; exit 0; }
      pyenv global ${PY_VER}
      python -m pip install --upgrade pip || true
    else
      echo 'Warning: pyenv not found; skipping pyenv-managed Python'
    fi
  "

  # rustup components
  sudo -u "$user" bash -lc '
    set -e
    [ -f "$HOME/.profile" ] && . "$HOME/.profile"
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
    export PATH="$HOME/.cargo/bin:$PATH"
    command -v rustup >/dev/null 2>&1 || exit 0
    rustup default stable || true
    rustup component add rustfmt clippy || true
  '

  # Podman basic smoke test (best-effort)
  sudo -u "$user" bash -lc '
    set -e
    if command -v podman >/dev/null 2>&1; then
      podman system migrate >/dev/null 2>&1 || true
      timeout 20s podman run --rm --network slirp4netns --dns 1.1.1.1 quay.io/podman/hello >/dev/null 2>&1 || true
    fi
  '

  # Versions summary
  sudo -u "$user" bash -lc '
    set -e
    [ -f "$HOME/.profile" ] && . "$HOME/.profile"
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
    export NVM_DIR="$HOME/.nvm"; [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
    export PYENV_ROOT="$HOME/.pyenv"; [ -d "$PYENV_ROOT/bin" ] && export PATH="$PYENV_ROOT/bin:$PATH"; command -v pyenv >/dev/null 2>&1 && eval "$(pyenv init -)"
    export PATH="$HOME/.cargo/bin:$PATH"
    echo "[*] Versions summary:"
    command -v node >/dev/null 2>&1 && echo "  node: $(node -v)" || true
    command -v python >/dev/null 2>&1 && echo "  python: $(python -V 2>&1)" || true
    command -v rustc >/dev/null 2>&1 && echo "  rustc: $(rustc -V 2>/dev/null)" || true
    command -v podman >/dev/null 2>&1 && echo "  podman: $(podman --version 2>/dev/null)" || true
  '
}

cleanup_for_snapshot() {
  log "*" "Cleaning caches and logs"
  yes | sudo pacman -Scc --noconfirm >/dev/null 2>&1 || true
  rm -rf "$HOME/.cache"/* 2>/dev/null || true
  sudo rm -rf /var/cache/pacman/pkg/* 2>/dev/null || true
  sudo journalctl --rotate >/dev/null 2>&1 || true
  sudo journalctl --vacuum-time=1s >/dev/null 2>&1 || true
}

# --------------------------------------------
# Phases
# --------------------------------------------
phase1_main() {
  log "*" "Starting Phase 1"
  log "i" "Config: user=${DEFAULT_USER}, DNS_MODE=${DNS_MODE}, OPTIMIZE_MIRRORS=${OPTIMIZE_MIRRORS}"
  ensure_locale
  pacman_quiet_update
  install_packages
  ensure_user_and_sudo "$DEFAULT_USER"
  ensure_subids "$DEFAULT_USER"
  configure_wsl
  configure_dns
  install_pyenv_nvm_rustup_for_user "$DEFAULT_USER"
  link_dotfiles_for_user "$DEFAULT_USER"
  log "+" "Phase 1 complete. Terminate WSL and start a new session for Phase 2."
}

phase2_main() {
  log "*" "Starting Phase 2"
  enable_services
  finalize_user_toolchains "$DEFAULT_USER"
  cleanup_for_snapshot
  if [ "$DNS_MODE" = "resolved" ]; then
    if command -v resolvectl >/dev/null 2>&1; then
      log "i" "systemd-resolved active. Check: resolvectl status"
    fi
    if [ -L /etc/resolv.conf ]; then
      log "i" "/etc/resolv.conf -> $(readlink -f /etc/resolv.conf)"
    fi
  fi
  log "+" "Phase 2 complete. You can terminate the distro and export a snapshot."
}

case "$PHASE" in
  phase1) phase1_main ;;
  phase2) phase2_main ;;
  *) echo "Usage: $0 [phase1|phase2]" >&2; exit 2 ;;
esac