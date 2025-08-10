# Dotfiles linking (library)
#
# Description: Safe symlink creation and linking of repository dotfiles into a
# user home directory.
# Globals: REPO_ROOT_MNT

#######################################
# Safely create/update a symlink
# Globals: None
# Args: $1 - src, $2 - dest
#######################################
safe_link() {
  local src="$1" dest="$2"
  mkdir -p "$(dirname "$dest")"
  if [[ -e "$dest" && ! -L "$dest" ]]; then
    mv -n "$dest" "${dest}.bak" || true
  fi
  ln -sfn "$src" "$dest"
}

#######################################
# Safely create/update a symlink as a specific user
# Globals: None
# Args: $1 - user, $2 - src, $3 - dest
#######################################
safe_link_user() {
  local user="$1" src="$2" dest="$3"
  sudo -u "$user" bash -lc "mkdir -p \"\$(dirname \"$dest\")\""
  if sudo -u "$user" test -e "$dest" && ! sudo -u "$user" test -L "$dest"; then
    sudo -u "$user" mv -n "$dest" "${dest}.bak" || true
  fi
  sudo -u "$user" ln -sfn "$src" "$dest"
}

#######################################
# Link repo dotfiles for a user
# Globals: REPO_ROOT_MNT
# Args: $1 - username
#######################################
link_dotfiles_for_user() {
  local user="$1"
  local home_dir; home_dir="$(eval echo ~"$user")"
  local dotroot=""

  if [[ -n "${REPO_ROOT_MNT:-}" && -d "${REPO_ROOT_MNT}/dotfiles" ]]; then
    dotroot="${REPO_ROOT_MNT}/dotfiles"
  else
    if command -v powershell.exe >/dev/null 2>&1; then
      local win_cwd
      win_cwd="$(powershell.exe -NoProfile -Command '$pwd.Path' | tr -d '\r')"
      if [[ -n "$win_cwd" ]]; then
        local repo_root
        repo_root="$(wslpath -u "$win_cwd")"
        [[ -d "${repo_root}/dotfiles" ]] && dotroot="${repo_root}/dotfiles"
      fi
    fi
  fi

  if [[ -n "$dotroot" && -d "$dotroot" ]]; then
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
