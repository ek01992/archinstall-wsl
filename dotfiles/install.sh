#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE}")" && pwd)"
echo "[*] Installing dotfiles from $ROOT"

mkdir -p "$HOME/.config" "$HOME/.cargo" "$HOME/.config/nvim"

link() {
  local src="$1" dest="$2"
  mkdir -p "$(dirname "$dest")"
  if [ -e "$dest" ] && [ ! -L "$dest" ]; then
    mv -n "$dest" "${dest}.bak" || true
  fi
  ln -sfn "$src" "$dest"
}

link "$ROOT/bash/.bashrc" "$HOME/.bashrc"
link "$ROOT/bash/.bash_aliases" "$HOME/.bash_aliases"
link "$ROOT/git/.gitconfig" "$HOME/.gitconfig"
link "$ROOT/nvim/init.vim" "$HOME/.config/nvim/init.vim"
link "$ROOT/rust/config.toml" "$HOME/.cargo/config.toml"
link "$ROOT/editorconfig/.editorconfig" "$HOME/.editorconfig"

echo "[*] Dotfiles linked."