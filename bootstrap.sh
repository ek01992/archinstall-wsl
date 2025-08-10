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
  base-devel sudo curl git wget unzip zip jq ripgrep fzf tmux htop lsof rsync openssh net-tools
  # Python build deps (ensure ssl, sqlite3, tkinter, etc. are compiled)
  openssl zlib xz tk readline sqlite gdbm libffi bzip2 ncurses
)
CONTAINER_PACKAGES=(
  crun conmon containers-common aardvark-dns netavark slirp4netns fuse-overlayfs podman buildah skopeo
)
OPTIONAL_LANG_PACKAGES=( go )
OPTIONAL_DOTFILES_HELPERS=( neovim direnv starship )
OPTIMIZE_MIRRORS="${OPTIMIZE_MIRRORS:-true}"

# --------------------------------------------
# Logging and small helpers
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

# --------------------------------------------
# TTY-aware UI helpers (colors, sections, spinner)
# --------------------------------------------
if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  BOLD="$(printf '\033[1m')"; RESET="$(printf '\033[0m')"
  CYAN="$(printf '\033[36m')"; GREEN="$(printf '\033[32m')"
  YELLOW="$(printf '\033[33m')"; RED="$(printf '\033[31m')"
else
  BOLD=""; RESET=""; CYAN=""; GREEN=""; YELLOW=""; RED=""
fi

ui::cols() { tput cols 2>/dev/null || printf 80; }
ui::hr() { local w; w="$(ui::cols)"; printf '%*s' "$(( w-2 ))" | tr ' ' '─'; }

ui::section() {
  printf "%b\n" "${CYAN}$(ui::hr)${RESET}"
  printf "%b\n" "${BOLD}${CYAN}▶ $*${RESET}"
  printf "%b\n" "${CYAN}$(ui::hr)${RESET}"
}
ui::info()  { printf "%b\n" "${CYAN}[*]${RESET} $*"; }
ui::ok()    { printf "%b\n" "${GREEN}[+]${RESET} $*"; }
ui::warn()  { printf "%b\n" "${YELLOW}[!]${RESET} $*"; }
ui::err()   { printf "%b\n" "${RED}[x]${RESET} $*"; }

# Spinner: ui::run_with_spinner "Message" cmd args...
ui::run_with_spinner() {
  local msg="$1"; shift
  local spin='-\|/'; local i=0
  "$@" & local pid=$!
  while kill -0 "$pid" 2>/dev/null; do
    i=$(( (i+1) %4 ))
    printf "\r${CYAN}[%s]${RESET} %s" "${spin:$i:1}" "$msg"
    sleep 0.1
  done
  wait "$pid"; local rc=$?
  printf "\r"
  if [ $rc -eq 0 ]; then ui::ok "$msg"; else ui::err "$msg (rc=$rc)"; fi
  return $rc
}

# Choice: gum > fzf > select
ui::choose() {
  local prompt="$1"; shift
  if command -v gum >/dev/null 2>&1; then
    gum choose "$@"
  elif command -v fzf >/dev/null 2>&1; then
    printf "%s\n" "$@" | fzf --prompt="${prompt} > "
  else
    PS3="${prompt}: "; select c in "$@"; do [ -n "${c:-}" ] && printf "%s" "$c" && break; done
  fi
}

# Provide sudo fallback when not installed (Phase 1 runs as root).
# Support "sudo -u ..." and plain "sudo ..."
if ! command -v sudo >/dev/null 2>&1; then
  sudo() {
    if [ "$#" -ge 3 ] && [ "$1" = "-u" ]; then
      shift
      local __sudo_user="$1"; shift
      local __cmd=""
      local __part
      for __part in "$@"; do
        __cmd+=$(printf ' %q' "$__part")
      done
      su -s /bin/bash - "$__sudo_user" -c "$__cmd"
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
  if sudo -u "$user" test -e "$dest" && ! sudo -u "$user" test -L "$dest"; then
    sudo -u "$user" mv -n "$dest" "${dest}.bak" || true
  fi
  sudo -u "$user" ln -sfn "$src" "$dest"
}

get_user_home() {
  local user="$1"
  local hd
  hd="$(getent passwd "$user" | cut -d: -f6 || true)"
  if [ -z "${hd:-}" ]; then
    hd="$(su -s /bin/bash - "$user" -c 'printf %s "$HOME"' 2>/dev/null || true)"
  fi
  [ -n "${hd:-}" ] && printf '%s\n' "$hd" || return 1
}

ensure_user_rc_files() {
  local user="$1"
  local home_dir
  home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"
  install -d -o "$user" -g "$user" "$home_dir"
  if [ ! -e "$home_dir/.bashrc" ]; then
    install -o "$user" -g "$user" -m 0644 -D /dev/null "$home_dir/.bashrc"
  else
    chown "$user:$user" "$home_dir/.bashrc" || true
  fi
  if [ ! -e "$home_dir/.profile" ]; then
    install -o "$user" -g "$user" -m 0644 -D /dev/null "$home_dir/.profile"
  else
    chown "$user:$user" "$home_dir/.profile" || true
  fi
}

# --------------------------------------------
# Locale configuration (decomposed)
# --------------------------------------------
configure_locale_gen() {
  ui::info "Configuring /etc/locale.gen for en_US.UTF-8"
  sudo sed -i 's/^# *en_US.UTF-8/en_US.UTF-8/' /etc/locale.gen
}

generate_locale() {
  ui::info "Generating locales"
  sudo locale-gen
}

persist_locale_env() {
  ui::info "Persisting locale environment"
  echo 'LANG=en_US.UTF-8' | sudo tee /etc/locale.conf >/dev/null
  append_once 'LANG=en_US.UTF-8' /etc/environment
  export LANG=en_US.UTF-8
  append_once 'export LANG=en_US.UTF-8' /root/.bashrc
}

ensure_locale() {
  configure_locale_gen
  generate_locale
  persist_locale_env
}

# --------------------------------------------
# Pacman configuration and update
# --------------------------------------------
ensure_pacman_parallel_downloads() {
  if ! grep -q '^\[options\]' /etc/pacman.conf; then
    printf "\n[options]\n" | sudo tee -a /etc/pacman.conf >/dev/null
  fi

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
}

install_reflector_best_effort() {
  ui::info "Installing reflector (best-effort)"
  sudo pacman -Sy --noconfirm --noprogressbar || true
  sudo pacman -S --needed --noconfirm --noprogressbar rsync reflector || true
}

update_mirrorlist_with_reflector() {
  if command -v reflector >/dev/null 2>&1; then
    if ! reflector --latest 20 --sort rate --save /etc/pacman.d/mirrorlist; then
      ui::warn "reflector failed; keeping existing mirrorlist"
    else
      ui::ok "Mirrorlist updated by reflector"
    fi
  fi
}

optimize_mirrors_if_enabled() {
  if [ "${OPTIMIZE_MIRRORS}" != "true" ]; then
    return 0
  fi
  ensure_pacman_parallel_downloads
  install_reflector_best_effort
  update_mirrorlist_with_reflector
}

refresh_pacman_keyring() {
  ui::info "Refreshing keyring"
  if command -v timedatectl >/dev/null 2>&1; then
    timedatectl status >/dev/null 2>&1 || true
  fi
  sudo rm -rf /etc/pacman.d/gnupg || true
  sudo pacman-key --init || true
  sudo pacman-key --populate archlinux || true
  retry 3 3 sudo pacman -Sy --noconfirm --noprogressbar archlinux-keyring || true
  sudo pacman-key --populate archlinux || true
}

pacman_system_update() {
  optimize_mirrors_if_enabled
  retry 3 5 sudo pacman -Syyu --noconfirm --noprogressbar
}

pacman_quiet_update() {
  ui::info "Updating base system (Syyu)"
  refresh_pacman_keyring
  pacman_system_update
}

install_packages() {
  ui::info "Installing packages"
  local pkgs=("${BASE_PACKAGES[@]}" "${CONTAINER_PACKAGES[@]}")
  retry 3 5 sudo pacman -S --needed --noconfirm --noprogressbar "${pkgs[@]}"
  ui::info "Optional language toolchains"
  sudo pacman -S --needed --noconfirm --noprogressbar "${OPTIONAL_LANG_PACKAGES[@]}" || true
  ui::info "Optional dotfiles helpers (best-effort)"
  sudo pacman -S --needed --noconfirm --noprogressbar "${OPTIONAL_DOTFILES_HELPERS[@]}" || true
}

# --------------------------------------------
# User and privileges
# --------------------------------------------
ensure_user_and_sudo() {
  local user="$1"
  if ! id -u "$user" >/dev/null 2>&1; then
    ui::info "Creating user $user with passwordless sudo"
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

# --------------------------------------------
# WSL configuration
# --------------------------------------------
configure_wsl() {
  ui::info "Writing /etc/wsl.conf (systemd=true, resolv.conf mode: ${DNS_MODE}, default user)"
  local gen="true"
  case "${DNS_MODE}" in
    static|resolved) gen="false" ;;
    wsl) gen="true" ;;
    *) ui::warn "Unknown DNS_MODE='${DNS_MODE}', defaulting to static"; gen="false"; DNS_MODE="static" ;;
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

# --------------------------------------------
# DNS configuration
# --------------------------------------------
ensure_resolv_unlocked() {
  if command -v chattr >/dev/null 2>&1; then
    sudo chattr -i /etc/resolv.conf 2>/dev/null || true
  fi
  sudo rm -f /etc/resolv.conf || true
}

write_static_resolv_conf() {
  {
    for ns in "${NAMESERVERS[@]}"; do
      printf 'nameserver %s\n' "$ns"
    done
    echo "options edns0"
  } | sudo tee /etc/resolv.conf >/dev/null
}

lock_resolv_if_enabled() {
  if [ "${CHATTR_IMMUTABLE_RESOLV}" = "true" ] && command -v chattr >/dev/null 2>&1; then
    sudo chattr +i /etc/resolv.conf || true
  fi
}

configure_dns_static() {
  ui::info "Configuring static resolv.conf"
  ensure_resolv_unlocked
  write_static_resolv_conf
  lock_resolv_if_enabled
}

create_resolved_link_service() {
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

configure_dns_resolved() {
  ui::info "Configuring systemd-resolved support"
  ensure_resolv_unlocked
  create_resolved_link_service
}

configure_dns_wsl() {
  ui::info "Using WSL-managed resolv.conf"
  ensure_resolv_unlocked
}

configure_dns() {
  case "${DNS_MODE}" in
    static)   configure_dns_static ;;
    resolved) configure_dns_resolved ;;
    wsl)      configure_dns_wsl ;;
    *)        ui::warn "Unknown DNS_MODE='${DNS_MODE}', defaulting to static"; DNS_MODE="static"; configure_dns_static ;;
  esac
}

# --------------------------------------------
# User toolchains and dotfiles
# --------------------------------------------
install_pyenv_for_user() {
  local user="$1"
  local home_dir
  home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"

  ui::info "Setting up pyenv for ${user}"
  ensure_user_rc_files "$user"

  if [ ! -d "${home_dir}/.pyenv" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc 'set -e; command -v curl >/dev/null 2>&1; curl -fsSL https://pyenv.run | bash'
  fi

  append_once 'export PYENV_ROOT="$HOME/.pyenv"' "${home_dir}/.bashrc"
  append_once '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' "${home_dir}/.bashrc"
  append_once 'eval "$(pyenv init -)"' "${home_dir}/.bashrc"
  append_once 'eval "$(pyenv virtualenv-init -)"' "${home_dir}/.bashrc"
  append_once 'export PYENV_ROOT="$HOME/.pyenv"' "${home_dir}/.profile"
  append_once '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' "${home_dir}/.profile"
  append_once 'eval "$(pyenv init -)"' "${home_dir}/.profile"

  sudo chown "$user:$user" "${home_dir}/.bashrc" "${home_dir}/.profile" 2>/dev/null || true
  sudo chown -R "$user:$user" "${home_dir}/.pyenv" 2>/dev/null || true
}

install_nvm_for_user() {
  local user="$1"
  local home_dir
  home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"

  ui::info "Setting up nvm for ${user}"
  ensure_user_rc_files "$user"

  if [ ! -d "${home_dir}/.nvm" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc 'set -e; command -v curl >/dev/null 2>&1; curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash'
  fi

  append_once 'export NVM_DIR="$HOME/.nvm"' "${home_dir}/.bashrc"
  append_once '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' "${home_dir}/.bashrc"
  append_once '[ -s "$NVM_DIR/bash_completion" ] && . "$NVM_DIR/bash_completion"' "${home_dir}/.bashrc"
  append_once 'export NVM_DIR="$HOME/.nvm"' "${home_dir}/.profile"
  append_once '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' "${home_dir}/.profile"

  sudo chown "$user:$user" "${home_dir}/.bashrc" "${home_dir}/.profile" 2>/dev/null || true
  sudo chown -R "$user:$user" "${home_dir}/.nvm" 2>/dev/null || true
}

install_rustup_for_user() {
  local user="$1"
  local home_dir
  home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"

  ui::info "Setting up rustup for ${user}"
  ensure_user_rc_files "$user"

  if [ ! -x "${home_dir}/.cargo/bin/rustc" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc 'set -e; command -v curl >/dev/null 2>&1; curl --proto "=https" --tlsv1.2 -fsSL https://sh.rustup.rs | sh -s -- -y'
  fi
  append_once 'export PATH="$HOME/.cargo/bin:$PATH"' "${home_dir}/.bashrc"

  sudo chown "$user:$user" "${home_dir}/.bashrc" "${home_dir}/.profile" 2>/dev/null || true
  sudo chown -R "$user:$user" "${home_dir}/.cargo" 2>/dev/null || true
}

detect_dotfiles_root() {
  local root=""
  if [ -n "$REPO_ROOT_MNT" ] && [ -d "$REPO_ROOT_MNT/dotfiles" ]; then
    root="${REPO_ROOT_MNT}/dotfiles"
  else
    if command -v powershell.exe >/dev/null 2>&1; then
      local win_cwd
      win_cwd="$(powershell.exe -NoProfile -Command '$pwd.Path' | tr -d '\r')"
      if [ -n "$win_cwd" ]; then
        local repo_root
        repo_root="$(wslpath -u "$win_cwd")"
        [ -d "${repo_root}/dotfiles" ] && root="${repo_root}/dotfiles"
      fi
    fi
  fi
  printf '%s' "$root"
}

link_dotfiles_for_user() {
  local user="$1"
  local home_dir
  home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"

  local dotroot
  dotroot="$(detect_dotfiles_root)"

  if [ -n "$dotroot" ] && [ -d "$dotroot" ]; then
    ui::ok "Linking dotfiles for ${user} from ${dotroot}"
    install -d -o "$user" -g "$user" "$home_dir/.config" "$home_dir/.cargo" "$home_dir/.config/nvim"
    safe_link_user "$user" "$dotroot/bash/.bashrc" "$home_dir/.bashrc"
    safe_link_user "$user" "$dotroot/bash/.bash_aliases" "$home_dir/.bash_aliases"
    safe_link_user "$user" "$dotroot/git/.gitconfig" "$home_dir/.gitconfig"
    safe_link_user "$user" "$dotroot/nvim/init.vim" "$home_dir/.config/nvim/init.vim"
    safe_link_user "$user" "$dotroot/rust/config.toml" "$home_dir/.cargo/config.toml"
    safe_link_user "$user" "$dotroot/editorconfig/.editorconfig" "$home_dir/.editorconfig"
  else
    ui::warn "dotfiles folder not found; skipped"
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

ensure_systemd_active_or_fail() {
  if ! is_systemd_active; then
    fail "systemd is not active (PID 1 != systemd). Restart WSL with systemd=true and rerun phase2."
  fi
}

configure_sshd_loopback() {
  echo 'ListenAddress 127.0.0.1' | sudo tee /etc/ssh/sshd_config.d/loopback.conf >/dev/null
}

enable_resolved_services_if_needed() {
  if [ "${DNS_MODE}" = "resolved" ]; then
    sudo systemctl enable --now systemd-resolved || true
    sudo systemctl enable --now wsl-resolved-link.service || true
  fi
}

enable_core_services() {
  sudo systemctl daemon-reload || true
  enable_resolved_services_if_needed
  sudo systemctl enable --now podman.socket
  sudo systemctl enable --now sshd
}

enable_services() {
  ui::info "Enabling services (requires systemd)"
  ensure_systemd_active_or_fail
  configure_sshd_loopback
  enable_core_services
}

finalize_pyenv_python() {
  local user="$1"
  local PY_VER="${PY_VER:-3.12.5}"
  sudo -u "$user" bash -lc "
    set -e
    [ -f \"\$HOME/.profile\" ] && . \"\$HOME/.profile\"
    [ -f \"\$HOME/.bashrc\" ] && . \"\$HOME/.bashrc\"
    export PYENV_ROOT=\"\$HOME/.pyenv\"
    [ -d \"\$PYENV_ROOT/bin\" ] && export PATH=\"\$PYENV_ROOT/bin:\$PATH\"
    if command -v pyenv >/dev/null 2>&1; then
      eval \"\$(pyenv init -)\"
      if pyenv versions --bare | grep -qx \"${PY_VER}\"; then
        if ! PYENV_VERSION=\"${PY_VER}\" python -c \"import ssl, sqlite3, tkinter\" >/dev/null 2>&1; then
          echo \"Detected incomplete Python ${PY_VER}; rebuilding with required libs...\"
          pyenv uninstall -f ${PY_VER}
        fi
      fi
      pyenv install -s ${PY_VER} || { echo 'Warning: pyenv failed to install ${PY_VER}, will use system python'; exit 0; }
      pyenv global ${PY_VER}
      python -m pip install --upgrade pip || true
    else
      echo 'Warning: pyenv not found; skipping pyenv-managed Python'
    fi
  "
}

finalize_rust() {
  local user="$1"
  sudo -u "$user" bash -lc '
    set -e
    [ -f "$HOME/.profile" ] && . "$HOME/.profile"
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
    export PATH="$HOME/.cargo/bin:$PATH"
    command -v rustup >/dev/null 2>&1 || exit 0
    rustup default stable || true
    rustup component add rustfmt clippy || true
  '
}

finalize_podman() {
  local user="$1"
  sudo -u "$user" bash -lc '
    set -e
    if command -v podman >/dev/null 2>&1; then
      podman system migrate >/dev/null 2>&1 || true
      timeout 20s podman run --rm --network slirp4netns --dns 1.1.1.1 quay.io/podman/hello >/dev/null 2>&1 || true
    fi
  '
}

finalize_node_lts_if_nvm_present() {
  local user="$1"
  sudo -u "$user" bash -lc '
    set -e
    [ -f "$HOME/.profile" ] && . "$HOME/.profile"
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] || exit 0
    . "$NVM_DIR/nvm.sh"
    nvm install --lts >/dev/null 2>&1 || true
    nvm use --lts >/dev/null 2>&1 || true
  '
}

print_versions_summary() {
  local user="$1"
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

finalize_user_toolchains() {
  local user="$1"
  ui::info "Finalizing toolchains for ${user}"
  finalize_pyenv_python "$user"
  finalize_rust "$user"
  finalize_podman "$user"
  finalize_node_lts_if_nvm_present "$user"
  print_versions_summary "$user"
}

cleanup_for_snapshot() {
  ui::info "Cleaning caches and logs"
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
  ui::section "Phase 1"
  ui::info "Config: user=${DEFAULT_USER}, DNS_MODE=${DNS_MODE}, OPTIMIZE_MIRRORS=${OPTIMIZE_MIRRORS}"

  # Optional interactive DNS selection if requested and in TTY
  if [[ -t 1 && "${PROMPT_DNS_MODE:-0}" = "1" ]]; then
    DNS_MODE="$(ui::choose "DNS mode" static resolved wsl)"
    ui::ok "Selected DNS mode: ${DNS_MODE}"
  fi

  ensure_locale
  ui::run_with_spinner "Updating system" pacman_quiet_update
  ui::run_with_spinner "Installing packages" install_packages
  ensure_user_and_sudo "$DEFAULT_USER"
  ensure_subids "$DEFAULT_USER"
  configure_wsl
  configure_dns

  install_pyenv_for_user "$DEFAULT_USER"
  install_nvm_for_user "$DEFAULT_USER"
  install_rustup_for_user "$DEFAULT_USER"

  link_dotfiles_for_user "$DEFAULT_USER"
  ui::ok "Phase 1 complete. Terminate WSL and start a new session for Phase 2."
}

phase2_main() {
  ui::section "Phase 2"
  enable_services
  finalize_user_toolchains "$DEFAULT_USER"
  cleanup_for_snapshot
  if [ "$DNS_MODE" = "resolved" ]; then
    if command -v resolvectl >/dev/null 2>&1; then
      ui::info "systemd-resolved active. Check: resolvectl status"
    fi
    if [ -L /etc/resolv.conf ]; then
      ui::info "/etc/resolv.conf -> $(readlink -f /etc/resolv.conf)"
    fi
  fi
  ui::ok "Phase 2 complete. You can terminate the distro and export a snapshot."
}

case "$PHASE" in
  phase1) phase1_main ;;
  phase2) phase2_main ;;
  *) echo "Usage: $0 [phase1|phase2]" >&2; exit 2 ;;
esac