# Developer toolchains (library)
#
# Description: Install and finalize pyenv, nvm, and rustup for a user.
# Globals: None

#######################################
# Install pyenv, nvm, rustup and ensure shell hooks
# Globals: None
# Args: $1 - username
#######################################
install_pyenv_nvm_rustup_for_user() {
  local user="$1"
  local home_dir; home_dir="$(eval echo ~"$user")"

  log "*" "Setting up pyenv, nvm, rustup for ${user}"

  if [[ ! -d "${home_dir}/.pyenv" ]]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc \
      'set -e; command -v curl >/dev/null 2>&1; curl -fsSL https://pyenv.run | bash'
  fi

  append_once 'export PYENV_ROOT="$HOME/.pyenv"' "${home_dir}/.bashrc"
  append_once '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' "${home_dir}/.bashrc"
  append_once 'eval "$(pyenv init -)"' "${home_dir}/.bashrc"
  append_once 'eval "$(pyenv virtualenv-init -)"' "${home_dir}/.bashrc"
  append_once 'export PYENV_ROOT="$HOME/.pyenv"' "${home_dir}/.profile"
  append_once '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' "${home_dir}/.profile"
  append_once 'eval "$(pyenv init -)"' "${home_dir}/.profile"

  if [[ ! -d "${home_dir}/.nvm" ]]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc \
      'set -e; command -v curl >/dev/null 2>&1; curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash'
  fi
  append_once 'export NVM_DIR="$HOME/.nvm"' "${home_dir}/.bashrc"
  append_once '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' "${home_dir}/.bashrc"
  append_once '[ -s "$NVM_DIR/bash_completion" ] && . "$NVM_DIR/bash_completion"' "${home_dir}/.bashrc"
  append_once 'export NVM_DIR="$HOME/.nvm"' "${home_dir}/.profile"
  append_once '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' "${home_dir}/.profile"

  if [[ ! -x "${home_dir}/.cargo/bin/rustc" ]]; then
    sudo -u "$user" env -i HOME="$home_dir" PATH="/usr/bin:/bin" bash -lc \
      'set -e; command -v curl >/dev/null 2>&1; curl --proto "=https" --tlsv1.2 -fsSL https://sh.rustup.rs | sh -s -- -y'
  fi
  append_once 'export PATH="$HOME/.cargo/bin:$PATH"' "${home_dir}/.bashrc"

  chown "$user:$user" "${home_dir}/.bashrc" "${home_dir}/.profile" 2>/dev/null || true
  chown -R "$user:$user" "${home_dir}/.pyenv" "${home_dir}/.nvm" "${home_dir}/.cargo" 2>/dev/null || true
}

#######################################
# Finalize toolchains: Node LTS, Python via pyenv, rust components
# Globals: None
# Args: $1 - username
#######################################
finalize_user_toolchains() {
  local user="$1"
  local PY_VER="${PY_VER:-3.12.5}"

  log "*" "Finalizing toolchains for ${user} (Node LTS, Python ${PY_VER}, rust components)"

  sudo -u "$user" bash -lc '
    set -e
    [ -f "$HOME/.profile" ] && . "$HOME/.profile"
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
    nvm install --lts
    nvm alias default lts/*
  '

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

  sudo -u "$user" bash -lc '
    set -e
    [ -f "$HOME/.profile" ] && . "$HOME/.profile"
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
    export PATH="$HOME/.cargo/bin:$PATH"
    command -v rustup >/dev/null 2>&1 || exit 0
    rustup default stable || true
    rustup component add rustfmt clippy || true
  '

  sudo -u "$user" bash -lc '
    set -e
    [ -f "$HOME/.profile" ] && . "$HOME/.profile"
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
    export NVM_DIR="$HOME/.nvm"; [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
    export PYENV_ROOT="$HOME/.pyenv"; [ -d "$PYENV_ROOT/bin" ] && export PATH="$PYENV_ROOT/bin:$PATH"; command -v pyenv >/dev/null 2>&1 && eval "$(pyenv init -)"
    export PATH="$HOME/.cargo/bin:$PATH"
    echo "[*] Versions summary:"
    command -v node   >/dev/null 2>&1 && echo "  node: $(node -v)" || true
    command -v python >/dev/null 2>&1 && echo "  python: $(python -V 2>&1)" || true
    command -v rustc  >/dev/null 2>&1 && echo "  rustc: $(rustc -V 2>/dev/null)" || true
    command -v podman >/dev/null 2>&1 && echo "  podman: $(podman --version 2>/dev/null)" || true
  '
}
