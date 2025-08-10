#!/usr/bin/env bash
set -euo pipefail
trap 'printf "%s\n" "[x] Error at ${BASH_SOURCE[0]}:${LINENO}: ${BASH_COMMAND}"' ERR

PHASE="${1:-phase1}"
DEFAULT_USER="${DEFAULT_USER:-erik}"
WSL_MEMORY="${WSL_MEMORY:-8GB}"
WSL_CPUS="${WSL_CPUS:-4}"
REPO_ROOT_MNT="${REPO_ROOT_MNT:-}"

# DNS strategy: static | resolved | wsl
DNS_MODE="${DNS_MODE:-static}"

# -----------------------------
# Declarative toggles and params
# -----------------------------
NAMESERVERS=("1.1.1.1" "9.9.9.9")
RESOLV_OPTIONS_CSV="${RESOLV_OPTIONS_CSV:-edns0}"
CHATTR_IMMUTABLE_RESOLV="${CHATTR_IMMUTABLE_RESOLV:-true}"

ALLOW_NOPASSWD_SUDO="${ALLOW_NOPASSWD_SUDO:-true}"
SUDO_NOPASSWD_CMDS="${SUDO_NOPASSWD_CMDS:-ALL}"

PYENV_VERSION_TAG="${PYENV_VERSION_TAG:-v2.4.17}"
NVM_VERSION_TAG="${NVM_VERSION_TAG:-v0.39.7}"
PY_VER="${PY_VER:-3.12.5}"
RUST_TOOLCHAIN="${RUST_TOOLCHAIN:-stable}"
RUST_COMPONENTS_CSV="${RUST_COMPONENTS_CSV:-rustfmt,clippy}"
NODE_VERSION="${NODE_VERSION:-}"
PYTHON_MODULES_CSV="${PYTHON_MODULES_CSV:-}"

OPTIMIZE_MIRRORS="${OPTIMIZE_MIRRORS:-true}"
ENABLE_CONTAINER_PKGS="${ENABLE_CONTAINER_PKGS:-true}"
ENABLE_OPTIONAL_LANG_PACKAGES="${ENABLE_OPTIONAL_LANG_PACKAGES:-true}"
ENABLE_OPTIONAL_DOTFILES_HELPERS="${ENABLE_OPTIONAL_DOTFILES_HELPERS:-true}"
ENABLE_PYENV="${ENABLE_PYENV:-true}"
ENABLE_PYENV_VIRTUALENV="${ENABLE_PYENV_VIRTUALENV:-true}"
PYENV_VIRTUALENV_TAG="${PYENV_VIRTUALENV_TAG:-}"
ENABLE_NVM="${ENABLE_NVM:-true}"
ENABLE_RUSTUP="${ENABLE_RUSTUP:-true}"
ENABLE_LINK_DOTFILES="${ENABLE_LINK_DOTFILES:-true}"
ENABLE_SSHD="${ENABLE_SSHD:-true}"
ENABLE_PODMAN_SOCKET="${ENABLE_PODMAN_SOCKET:-true}"
ENABLE_SYSTEMD_RESOLVED="${ENABLE_SYSTEMD_RESOLVED:-true}"
RUN_PODMAN_TEST="${RUN_PODMAN_TEST:-true}"

PHASE1_TASKS_CSV="${PHASE1_TASKS_CSV:-}"
PHASE2_TASKS_CSV="${PHASE2_TASKS_CSV:-}"

BASE_PACKAGES=(
  base-devel sudo curl git wget unzip zip jq ripgrep fzf tmux htop lsof rsync openssh net-tools
  openssl zlib xz tk readline sqlite gdbm libffi bzip2 ncurses
)
CONTAINER_PACKAGES=(
  crun conmon containers-common aardvark-dns netavark slirp4netns fuse-overlayfs podman buildah skopeo
)
OPTIONAL_LANG_PACKAGES=( go )
OPTIONAL_DOTFILES_HELPERS=( neovim direnv starship )

BASE_PACKAGES_CSV="${BASE_PACKAGES_CSV:-}"
CONTAINER_PACKAGES_CSV="${CONTAINER_PACKAGES_CSV:-}"
OPTIONAL_LANG_PACKAGES_CSV="${OPTIONAL_LANG_PACKAGES_CSV:-}"
OPTIONAL_DOTFILES_HELPERS_CSV="${OPTIONAL_DOTFILES_HELPERS_CSV:-}"

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

csv_to_array() {
  local csv="$1"; local -n out="$2"
  IFS=',' read -ra __parts <<< "$csv"
  out=()
  local p
  for p in "${__parts[@]}"; do
    p="${p#"${p%%[![:space:]]*}"}"
    p="${p%"${p##*[![:space:]]}"}"
    [ -n "$p" ] && out+=("$p")
  done
}

if [ -n "${NAMESERVERS_CSV:-}" ]; then csv_to_array "$NAMESERVERS_CSV" NAMESERVERS; fi
if [ -n "${BASE_PACKAGES_CSV:-}" ]; then csv_to_array "$BASE_PACKAGES_CSV" BASE_PACKAGES; fi
if [ -n "${CONTAINER_PACKAGES_CSV:-}" ]; then csv_to_array "$CONTAINER_PACKAGES_CSV" CONTAINER_PACKAGES; fi
if [ -n "${OPTIONAL_LANG_PACKAGES_CSV:-}" ]; then csv_to_array "$OPTIONAL_LANG_PACKAGES_CSV" OPTIONAL_LANG_PACKAGES; fi
if [ -n "${OPTIONAL_DOTFILES_HELPERS_CSV:-}" ]; then csv_to_array "$OPTIONAL_DOTFILES_HELPERS_CSV" OPTIONAL_DOTFILES_HELPERS; fi

_ui_is_tty() { [ -t 1 ]; }
_ui_cols() { tput cols 2>/dev/null || printf 80; }
_ui_utf8() { case "${LANG:-}${LC_ALL:-}${LC_CTYPE:-}" in *UTF-8*|*utf8*) return 0;; *) return 1;; esac; }

if _ui_is_tty && [ -z "${NO_COLOR:-}" ]; then
  BOLD="$(printf '\033[1m')"; RESET="$(printf '\033[0m')"
  CYAN="$(printf '\033[36m')"; GREEN="$(printf '\033[32m')"
  YELLOW="$(printf '\033[33m')"; RED="$(printf '\033[31m')"
  MUTED="$(printf '\033[90m')"
else
  BOLD=""; RESET=""; CYAN=""; GREEN=""; YELLOW=""; RED=""; MUTED=""
fi

if [ -n "${UI_ASCII:-}" ] || ! _ui_utf8; then
  UI_CHAR_TL="+"
  UI_CHAR_TR="+"
  UI_CHAR_BL="+"
  UI_CHAR_BR="+"
  UI_CHAR_H="-"
  UI_CHAR_V="|"
  # shellcheck disable=SC2034
  UI_LEADER=">"
else
  UI_CHAR_TL="┌"
  UI_CHAR_TR="┐"
  UI_CHAR_BL="└"
  UI_CHAR_BR="┘"
  UI_CHAR_H="─"
  UI_CHAR_V="│"
  # shellcheck disable=SC2034
  UI_LEADER="▶"
fi

UI_VERBOSE="${UI_VERBOSE:-0}"
UI_LOGFILE="${UI_LOGFILE:-}"

ui::cols() { _ui_cols; }
ui::hr() {
  local w; w="$(_ui_cols)"
  local n=$(( w - 2 )); [ "$n" -lt 10 ] && n=10
  printf '%*s' "$n" "" | tr ' ' "$UI_CHAR_H"
}
ui::section() { printf "%b%s%s%b\n" "${MUTED}" "${UI_CHAR_TL}$(ui::hr)${UI_CHAR_TR}" "" "${RESET}"; printf "%b%s %s%b\n" "${BOLD}${CYAN}" "${UI_CHAR_V}" "$*" "${RESET}"; printf "%b%s%s%b\n" "${MUTED}" "${UI_CHAR_BL}$(ui::hr)${UI_CHAR_BR}" "" "${RESET}"; }
ui::info()  { printf "%b %s\n" "${CYAN}[*]${RESET}" "$*"; }
ui::ok()    { printf "%b %s\n" "${GREEN}[+]${RESET}" "$*"; }
ui::warn()  { printf "%b %s\n" "${YELLOW}[!]${RESET}" "$*"; }
ui::err()   { printf "%b %s\n" "${RED}[x]${RESET}" "$*"; }

ui::run_with_spinner() {
  local msg="$1"; shift
  local spin='-\|/'; local i=0
  if [ -n "${UI_LOGFILE:-}" ]; then "$@" >>"$UI_LOGFILE" 2>&1 & local pid=$!; else "$@" >/dev/null 2>&1 & local pid=$!; fi
  local start; start="$(date +%s)"
  while kill -0 "$pid" 2>/dev/null; do i=$(( (i+1) %4 )); printf "\r${CYAN}[%s]${RESET} %s" "${spin:$i:1}" "$msg"; sleep 0.1; done
  wait "$pid"; local rc=$?; local ts=$(( $(date +%s) - start )); printf "\r"
  if [ "$rc" -eq 0 ]; then ui::ok "$msg (${ts}s)"; else ui::err "$msg (rc=$rc, ${ts}s)"; fi
  return "$rc"
}

ui::step() {
  local title="$1"; shift
  if [ "${UI_VERBOSE}" = "0" ] && [ -n "${UI_LOGFILE:-}" ]; then ui::run_with_spinner "$title" "$@"
  else
    ui::info "$title"; local start rc; start="$(date +%s)"
    if [ -n "${UI_LOGFILE:-}" ]; then "$@" 2>&1 | tee -a "$UI_LOGFILE"; rc=${PIPESTATUS[0]}; else "$@"; rc=$?; fi
    local ts=$(( $(date +%s) - start ))
    if [ "$rc" -eq 0 ]; then ui::ok "$title (${ts}s)"; else ui::err "$title failed (rc=$rc, ${ts}s)"; fi
    return "$rc"
  fi
}

if ! command -v sudo >/dev/null 2>&1; then
  sudo() {
    if [ "$#" -ge 3 ] && [ "$1" = "-u" ]; then
      shift; local __sudo_user="$1"; shift
      local __cmd=""; local __part; for __part in "$@"; do __cmd+=$(printf ' %q' "$__part"); done
      su -s /bin/bash - "$__sudo_user" -c "$__cmd"
    else "$@"; fi
  }
fi

append_once() {
  local line="$1" file="$2"
  mkdir -p "$(dirname "$file")"
  if [ ! -f "$file" ]; then printf '%s\n' "$line" >"$file"; return 0; fi
  if ! grep -qxF -- "$line" "$file" 2>/dev/null; then printf '%s\n' "$line" >>"$file"; fi
}

safe_link() {
  local src="$1" dest="$2"
  mkdir -p "$(dirname "$dest")"
  if [ -e "$dest" ] && [ ! -L "$dest" ]; then mv -n "$dest" "${dest}.bak" || true; fi
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
  local hd; hd="$(getent passwd "$user" | cut -d: -f6 || true)"
  if [ -z "${hd:-}" ]; then hd="$(su -s /bin/bash - "$user" -c "printf %s '$HOME'" 2>/dev/null || true)"; fi
  [ -n "${hd:-}" ] && printf '%s\n' "$hd" || return 1
}

ensure_user_rc_files() {
  local user="$1"
  local home_dir; home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"
  install -d -o "$user" -g "$user" "$home_dir"
  if [ ! -e "$home_dir/.bashrc" ]; then install -o "$user" -g "$user" -m 0644 -D /dev/null "$home_dir/.bashrc"; else chown "$user:$user" "$home_dir/.bashrc" || true; fi
  if [ ! -e "$home_dir/.profile" ]; then install -o "$user" -g "$user" -m 0644 -D /dev/null "$home_dir/.profile"; else chown "$user:$user" "$home_dir/.profile" || true; fi
}

configure_locale_gen() { ui::info "Configuring /etc/locale.gen for en_US.UTF-8"; sudo sed -i 's/^# *en_US.UTF-8/en_US.UTF-8/' /etc/locale.gen; }
generate_locale() { ui::info "Generating locales"; sudo locale-gen; }
persist_locale_env() {
  ui::info "Persisting locale environment"
  echo 'LANG=en_US.UTF-8' | sudo tee /etc/locale.conf >/dev/null
  # shellcheck disable=SC2016
  append_once 'LANG=en_US.UTF-8' /etc/environment
  export LANG=en_US.UTF-8
  # shellcheck disable=SC2016
  append_once 'export LANG=en_US.UTF-8' /root/.bashrc
}
ensure_locale() { configure_locale_gen; generate_locale; persist_locale_env; }

ensure_pacman_parallel_downloads() {
  if ! grep -q '^\[options\]' /etc/pacman.conf; then printf "\n[options]\n" | sudo tee -a /etc/pacman.conf >/dev/null; fi
  local pd; pd="$(nproc 2>/dev/null || echo 2)"; pd=$(( pd * 2 )); [ "$pd" -gt 16 ] && pd=16; [ "$pd" -lt 2 ] && pd=2
  sudo awk -v pd="$pd" '
    BEGIN{inopt=0; inserted=0}
    /^\[/{ inopt = ($0 ~ /^\[options\]/); print; next }
    {
      if (inopt && $1 ~ /^ParallelDownloads/) next
      print
      if (inopt && !inserted) { print "ParallelDownloads = " pd; inserted=1 }
    }
  ' /etc/pacman.conf | sudo tee /etc/pacman.conf.tmp >/dev/null
  sudo mv /etc/pacman.conf.tmp /etc/pacman.conf
}

install_reflector_best_effort() { ui::info "Installing reflector (best-effort)"; sudo pacman -Sy --noconfirm --noprogressbar || true; sudo pacman -S --needed --noconfirm --noprogressbar rsync reflector || true; }
update_mirrorlist_with_reflector() {
  if command -v reflector >/dev/null 2>&1; then
    if ! reflector --latest 20 --sort rate --save /etc/pacman.d/mirrorlist; then ui::warn "reflector failed; keeping existing mirrorlist"; else ui::ok "Mirrorlist updated by reflector"; fi
  fi
}
optimize_mirrors_if_enabled() { [ "${OPTIMIZE_MIRRORS}" != "true" ] && return 0; ensure_pacman_parallel_downloads; install_reflector_best_effort; update_mirrorlist_with_reflector; }

refresh_pacman_keyring() {
  ui::info "Refreshing keyring"
  if command -v timedatectl >/dev/null 2>&1; then timedatectl status >/dev/null 2>&1 || true; fi
  sudo rm -rf /etc/pacman.d/gnupg || true
  sudo pacman-key --init || true
  sudo pacman-key --populate archlinux || true
  retry 3 3 sudo pacman -Sy --noconfirm --noprogressbar archlinux-keyring || true
  sudo pacman-key --populate archlinux || true
}

pacman_system_update() { optimize_mirrors_if_enabled; retry 3 5 sudo pacman -Syyu --noconfirm --noprogressbar; }
pacman_quiet_update() { ui::info "Updating base system (Syyu)"; refresh_pacman_keyring; pacman_system_update; }

install_packages() {
  ui::info "Installing packages"
  local pkgs=("${BASE_PACKAGES[@]}")
  if [ "${ENABLE_CONTAINER_PKGS}" = "true" ]; then pkgs+=("${CONTAINER_PACKAGES[@]}"); fi
  retry 3 5 sudo pacman -S --needed --noconfirm --noprogressbar "${pkgs[@]}"
  if [ "${ENABLE_OPTIONAL_LANG_PACKAGES}" = "true" ]; then
    ui::info "Optional language toolchains"
    sudo pacman -S --needed --noconfirm --noprogressbar "${OPTIONAL_LANG_PACKAGES[@]}" || true
  fi
  if [ "${ENABLE_OPTIONAL_DOTFILES_HELPERS}" = "true" ]; then
    ui::info "Optional dotfiles helpers (best-effort)"
    sudo pacman -S --needed --noconfirm --noprogressbar "${OPTIONAL_DOTFILES_HELPERS[@]}" || true
  fi
}

ensure_user_and_sudo() {
  local user="$1"
  if ! id -u "$user" >/dev/null 2>&1; then ui::info "Creating user $user"; sudo useradd -m -s /bin/bash "$user"; fi
  if [ "${ALLOW_NOPASSWD_SUDO}" = "true" ]; then
    local sudo_file="/etc/sudoers.d/99-${user}-nopasswd"
    echo "${user} ALL=(ALL) NOPASSWD:${SUDO_NOPASSWD_CMDS}" | sudo tee "$sudo_file" >/dev/null
    sudo chmod 0440 "$sudo_file"
  else
    ui::warn "NOPASSWD sudo disabled via ALLOW_NOPASSWD_SUDO=false"
  fi
}

ensure_subids() {
  local user="$1"
  if ! grep -q "^${user}:" /etc/subuid 2>/dev/null; then
    sudo usermod --add-subuids 524288-589823 "$user" 2>/dev/null || echo "${user}:524288:65536" | sudo tee -a /etc/subuid >/dev/null
  fi
  if ! grep -q "^${user}:" /etc/subgid 2>/dev/null; then
    sudo usermod --add-subgids 524288-589823 "$user" 2>/dev/null || echo "${user}:524288:65536" | sudo tee -a /etc/subgid >/dev/null
  fi
}

configure_wsl() {
  ui::info "Writing /etc/wsl.conf (systemd=true, resolv.conf mode: ${DNS_MODE}, default user)"
  local gen="true"
  case "${DNS_MODE}" in static|resolved) gen="false" ;; wsl) gen="true" ;; *) ui::warn "Unknown DNS_MODE='${DNS_MODE}', defaulting to static"; gen="false"; DNS_MODE="static" ;; esac
  cat <<EOF | sudo tee /etc/wsl.conf >/dev/null
[boot]
systemd=true

[user]
default=${DEFAULT_USER}

[network]
generateResolvConf=${gen}
EOF
}

ensure_resolv_unlocked() { if command -v chattr >/dev/null 2>&1; then sudo chattr -i /etc/resolv.conf 2>/dev/null || true; fi; sudo rm -f /etc/resolv.conf || true; }

resolv_options_line() {
  [ -z "${RESOLV_OPTIONS_CSV:-}" ] && return 0
  IFS=',' read -ra __opts <<< "$RESOLV_OPTIONS_CSV"
  local opts_line="options"; local o
  for o in "${__opts[@]}"; do
    o="${o#"${o%%[![:space:]]*}"}"; o="${o%"${o##*[![:space:]]}"}"
    [ -n "$o" ] && opts_line="$opts_line $o"
  done
  printf '%s\n' "$opts_line"
}

write_static_resolv_conf() {
  {
    local ns
    for ns in "${NAMESERVERS[@]}"; do printf 'nameserver %s\n' "$ns"; done
    resolv_options_line || true
  } | sudo tee /etc/resolv.conf >/dev/null
}

lock_resolv_if_enabled() { if [ "${CHATTR_IMMUTABLE_RESOLV}" = "true" ] && command -v chattr >/dev/null 2>&1; then sudo chattr +i /etc/resolv.conf || true; fi; }

configure_dns_static() { ui::info "Configuring static resolv.conf"; ensure_resolv_unlocked; write_static_resolv_conf; lock_resolv_if_enabled; }

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

configure_dns_resolved() { ui::info "Configuring systemd-resolved support"; ensure_resolv_unlocked; create_resolved_link_service; }
configure_dns_wsl() { ui::info "Using WSL-managed resolv.conf"; ensure_resolv_unlocked; }

configure_dns() {
  case "${DNS_MODE}" in
    static)   configure_dns_static ;;
    resolved) configure_dns_resolved ;;
    wsl)      configure_dns_wsl ;;
    *)        ui::warn "Unknown DNS_MODE='${DNS_MODE}', defaulting to static"; DNS_MODE="static"; configure_dns_static ;;
  esac
}

install_pyenv_for_user() {
  [ "${ENABLE_PYENV}" != "true" ] && { ui::warn "pyenv disabled (ENABLE_PYENV=false)"; return 0; }
  local user="$1"; local home_dir; home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"
  ui::info "Setting up pyenv for ${user}"
  ensure_user_rc_files "$user"
  if [ ! -d "${home_dir}/.pyenv" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc "
      set -euo pipefail
      git clone https://github.com/pyenv/pyenv.git \"\$HOME/.pyenv\"
      cd \"\$HOME/.pyenv\"
      git checkout ${PYENV_VERSION_TAG}
    "
  fi
  if [ "${ENABLE_PYENV_VIRTUALENV}" = "true" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc "
      set -euo pipefail
      if [ ! -d \"\$HOME/.pyenv/plugins/pyenv-virtualenv\" ]; then
        git clone https://github.com/pyenv/pyenv-virtualenv.git \"\$HOME/.pyenv/plugins/pyenv-virtualenv\"
        if [ -n \"${PYENV_VIRTUALENV_TAG}\" ]; then
          cd \"\$HOME/.pyenv/plugins/pyenv-virtualenv\" && git checkout \"${PYENV_VIRTUALENV_TAG}\" || true
        fi
      fi
    "
  fi

  # shellcheck disable=SC2016
  append_once 'export PYENV_ROOT="$HOME/.pyenv"' "${home_dir}/.bashrc"
  # shellcheck disable=SC2016
  append_once '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' "${home_dir}/.bashrc"
  # shellcheck disable=SC2016
  append_once 'eval "$(pyenv init -)"' "${home_dir}/.bashrc"
  # shellcheck disable=SC2016
  append_once '[ -d "$HOME/.pyenv/plugins/pyenv-virtualenv" ] && eval "$(pyenv virtualenv-init -)"' "${home_dir}/.bashrc"
  # shellcheck disable=SC2016
  append_once 'export PYENV_ROOT="$HOME/.pyenv"' "${home_dir}/.profile"
  # shellcheck disable=SC2016
  append_once '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' "${home_dir}/.profile"
  # shellcheck disable=SC2016
  append_once 'eval "$(pyenv init -)"' "${home_dir}/.profile"

  sudo chown "$user:$user" "${home_dir}/.bashrc" "${home_dir}/.profile" 2>/dev/null || true
  sudo chown -R "$user:$user" "${home_dir}/.pyenv" 2>/dev/null || true
}

install_nvm_for_user() {
  [ "${ENABLE_NVM}" != "true" ] && { ui::warn "nvm disabled (ENABLE_NVM=false)"; return 0; }
  local user="$1"; local home_dir; home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"
  ui::info "Setting up nvm for ${user}"
  ensure_user_rc_files "$user"
  if [ ! -d "${home_dir}/.nvm" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc "
      set -euo pipefail
      git clone https://github.com/nvm-sh/nvm.git \"\$HOME/.nvm\"
      cd \"\$HOME/.nvm\"
      git checkout ${NVM_VERSION_TAG}
    "
  fi
  # shellcheck disable=SC2016
  append_once 'export NVM_DIR="$HOME/.nvm"' "${home_dir}/.bashrc"
  # shellcheck disable=SC2016
  append_once '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' "${home_dir}/.bashrc"
  # shellcheck disable=SC2016
  append_once '[ -s "$NVM_DIR/bash_completion" ] && . "$NVM_DIR/bash_completion"' "${home_dir}/.bashrc"
  # shellcheck disable=SC2016
  append_once 'export NVM_DIR="$HOME/.nvm"' "${home_dir}/.profile"
  # shellcheck disable=SC2016
  append_once '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' "${home_dir}/.profile"
  sudo chown "$user:$user" "${home_dir}/.bashrc" "${home_dir}/.profile" 2>/dev/null || true
  sudo chown -R "$user:$user" "${home_dir}/.nvm" 2>/dev/null || true
}

install_rustup_for_user() {
  [ "${ENABLE_RUSTUP}" != "true" ] && { ui::warn "rustup disabled (ENABLE_RUSTUP=false)"; return 0; }
  local user="$1"; local home_dir; home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"
  ui::info "Setting up rustup for ${user}"
  ensure_user_rc_files "$user"
  if [ ! -x "${home_dir}/.cargo/bin/rustc" ]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc 'set -e
      command -v curl >/dev/null 2>&1
      curl --proto "=https" --tlsv1.2 -fsSL https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain stable --no-modify-path
    '
  fi
  # shellcheck disable=SC2016
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
      local win_cwd; win_cwd="$(powershell.exe -NoProfile -Command "\$pwd.Path" | tr -d '\r')"
      if [ -n "$win_cwd" ]; then
        local repo_root; repo_root="$(wslpath -u "$win_cwd")"
        [ -d "${repo_root}/dotfiles" ] && root="${repo_root}/dotfiles"
      fi
    fi
  fi
  printf '%s' "$root"
}

link_dotfiles_for_user() {
  [ "${ENABLE_LINK_DOTFILES}" != "true" ] && { ui::warn "Link dotfiles disabled (ENABLE_LINK_DOTFILES=false)"; return 0; }
  local user="$1"; local home_dir; home_dir="$(get_user_home "$user")" || fail "Could not determine home for user '$user'"
  local dotroot; dotroot="$(detect_dotfiles_root)"
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

is_systemd_active() { local pid1; pid1="$(ps -p 1 -o comm= 2>/dev/null || true)"; [ "$pid1" = "systemd" ]; }
ensure_systemd_active_or_fail() { if ! is_systemd_active; then fail "systemd is not active (PID 1 != systemd). Restart WSL with systemd=true and rerun phase2."; fi; }

configure_sshd_loopback() { echo 'ListenAddress 127.0.0.1' | sudo tee /etc/ssh/sshd_config.d/loopback.conf >/dev/null; }

enable_resolved_services_if_needed() {
  if [ "${DNS_MODE}" = "resolved" ] && [ "${ENABLE_SYSTEMD_RESOLVED}" = "true" ]; then
    sudo systemctl enable --now systemd-resolved || true
    sudo systemctl enable --now wsl-resolved-link.service || true
  fi
}

enable_core_services() {
  sudo systemctl daemon-reload || true
  enable_resolved_services_if_needed
  if [ "${ENABLE_PODMAN_SOCKET}" = "true" ]; then sudo systemctl enable --now podman.socket; fi
  if [ "${ENABLE_SSHD}" = "true" ]; then sudo systemctl enable --now sshd; fi
}

enable_services() {
  ui::info "Enabling services (requires systemd)"
  ensure_systemd_active_or_fail
  configure_sshd_loopback
  enable_core_services
}

finalize_pyenv_python() {
  local user="$1"
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
      if [ -n \"${PYTHON_MODULES_CSV}\" ]; then
        IFS=',' read -ra __mods <<< \"${PYTHON_MODULES_CSV}\"
        python -m pip install -q --no-input \"\${__mods[@]}\" || true
      fi
    else
      echo 'Warning: pyenv not found; skipping pyenv-managed Python'
    fi
  "
}

finalize_rust() {
  local user="$1"
  sudo -u "$user" bash -lc "
    set -e
    [ -f \"\$HOME/.profile\" ] && . \"\$HOME/.profile\"
    [ -f \"\$HOME/.bashrc\" ] && . \"\$HOME/.bashrc\"
    export PATH=\"\$HOME/.cargo/bin:\$PATH\"
    command -v rustup >/dev/null 2>&1 || exit 0
    rustup default ${RUST_TOOLCHAIN} || true
    IFS=',' read -ra __comps <<< \"${RUST_COMPONENTS_CSV}\"
    for __c in \"\${__comps[@]}\"; do
      __c=\"\$(echo \"\$__c\" | sed 's/^[[:space:]]*//;s/[[:space:]]*\$//')\"
      [ -n \"\$__c\" ] && rustup component add \"\$__c\" || true
    done
  "
}

finalize_podman() {
  local user="$1"
  sudo -u "$user" bash -lc "
    set -e
    if command -v podman >/dev/null 2>&1; then
      podman system migrate >/dev/null 2>&1 || true
      if [ \"${RUN_PODMAN_TEST}\" = \"true\" ]; then
        timeout 20s podman run --rm --network slirp4netns --dns 1.1.1.1 quay.io/podman/hello >/dev/null 2>&1 || true
      fi
    fi
  "
}

finalize_node_if_nvm_present() {
  local user="$1"
  # shellcheck disable=SC2016
  sudo -u "$user" bash -lc "
    set -e
    [ -f \"\$HOME/.profile\" ] && . \"\$HOME/.profile\"
    [ -f \"\$HOME/.bashrc\" ] && . \"\$HOME/.bashrc\"
    export NVM_DIR=\"\$HOME/.nvm\"
    [ -s \"\$NVM_DIR/nvm.sh\" ] || exit 0
    . \"\$NVM_DIR/nvm.sh\"
    if [ -n \"${NODE_VERSION}\" ]; then
      nvm install \"${NODE_VERSION}\" >/dev/null 2>&1 || true
      nvm use \"${NODE_VERSION}\" >/dev/null 2>&1 || true
    else
      nvm install --lts >/dev/null 2>&1 || true
      nvm use --lts >/dev/null 2>&1 || true
    fi
  "
}

print_versions_summary() {
  local user="$1"
  # shellcheck disable=SC2016
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
  finalize_node_if_nvm_present "$user"
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

execute_phase1_task() {
  local t="$1"
  case "$t" in
    ensure_locale)                ui::step "Ensure locale" ensure_locale ;;
    pacman_update)                ui::step "Updating system" pacman_quiet_update ;;
    install_packages)             ui::step "Installing packages" install_packages ;;
    ensure_user_and_sudo)         ui::step "Ensuring user and sudo" ensure_user_and_sudo "$DEFAULT_USER" ;;
    ensure_subids)                ui::step "Allocating subordinate IDs" ensure_subids "$DEFAULT_USER" ;;
    configure_wsl)                ui::step "Write WSL config" configure_wsl ;;
    configure_dns)                ui::step "Configure DNS (${DNS_MODE})" configure_dns ;;
    install_pyenv)                if [ "${ENABLE_PYENV}" = "true" ]; then ui::step "Setup pyenv for ${DEFAULT_USER}" install_pyenv_for_user "$DEFAULT_USER"; else ui::warn "Skipped install_pyenv (disabled)"; fi ;;
    install_nvm)                  if [ "${ENABLE_NVM}" = "true" ]; then ui::step "Setup nvm for ${DEFAULT_USER}" install_nvm_for_user "$DEFAULT_USER"; else ui::warn "Skipped install_nvm (disabled)"; fi ;;
    install_rustup)               if [ "${ENABLE_RUSTUP}" = "true" ]; then ui::step "Setup rustup for ${DEFAULT_USER}" install_rustup_for_user "$DEFAULT_USER"; else ui::warn "Skipped install_rustup (disabled)"; fi ;;
    link_dotfiles)                if [ "${ENABLE_LINK_DOTFILES}" = "true" ]; then ui::step "Link dotfiles for ${DEFAULT_USER}" link_dotfiles_for_user "$DEFAULT_USER"; else ui::warn "Skipped link_dotfiles (disabled)"; fi ;;
    *) ui::warn "Unknown phase1 task: $t (skipped)";;
  esac
}

execute_phase2_task() {
  local t="$1"
  case "$t" in
    enable_services)              ui::step "Enable services" enable_services ;;
    finalize_user_toolchains)     ui::step "Finalize user toolchains" finalize_user_toolchains "$DEFAULT_USER" ;;
    cleanup_for_snapshot)         ui::step "Cleanup for snapshot" cleanup_for_snapshot ;;
    *) ui::warn "Unknown phase2 task: $t (skipped)";;
  esac
}

run_csv_tasks() {
  local csv="$1" phase="$2"
  local -a tasks
  csv_to_array "$csv" tasks
  local t
  for t in "${tasks[@]}"; do
    case "$phase" in
      1) execute_phase1_task "$t" ;;
      2) execute_phase2_task "$t" ;;
    esac
  done
}

phase1_main() {
  ui::section "Phase 1"
  ui::section "Preflight summary"
  ui::kv_table \
    "Default user" "${DEFAULT_USER}" \
    "DNS mode" "${DNS_MODE}" \
    "WSL CPUs" "${WSL_CPUS}" \
    "WSL memory" "${WSL_MEMORY}" \
    "Optimize mirrors" "${OPTIMIZE_MIRRORS}" \
    "Repo root (mnt)" "${REPO_ROOT_MNT:-<auto>}"

  if [ -n "${PHASE1_TASKS_CSV}" ]; then
    run_csv_tasks "$PHASE1_TASKS_CSV" 1
  else
    execute_phase1_task ensure_locale
    execute_phase1_task pacman_update
    execute_phase1_task install_packages
    execute_phase1_task ensure_user_and_sudo
    execute_phase1_task ensure_subids
    execute_phase1_task configure_wsl
    execute_phase1_task configure_dns
    execute_phase1_task install_pyenv
    execute_phase1_task install_nvm
    execute_phase1_task install_rustup
    execute_phase1_task link_dotfiles
  fi
  ui::ok "Phase 1 complete. Terminate WSL and start a new session for Phase 2."
}

phase2_main() {
  ui::section "Phase 2"
  if [ -n "${PHASE2_TASKS_CSV}" ]; then
    run_csv_tasks "$PHASE2_TASKS_CSV" 2
  else
    execute_phase2_task enable_services
    execute_phase2_task finalize_user_toolchains
    execute_phase2_task cleanup_for_snapshot
  fi
  if [ "$DNS_MODE" = "resolved" ]; then
    if command -v resolvectl >/dev/null 2>&1; then ui::info "systemd-resolved active. Check: resolvectl status"; fi
    if [ -L /etc/resolv.conf ]; then ui::info "/etc/resolv.conf -> $(readlink -f /etc/resolv.conf)"; fi
  fi
  ui::ok "Phase 2 complete. You can terminate the distro and export a snapshot."
}

case "$PHASE" in
  phase1) phase1_main ;;
  phase2) phase2_main ;;
  *) echo "Usage: $0 [phase1|phase2]" >&2; exit 2 ;;
esac