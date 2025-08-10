#!/bin/bash
set -Eeuo pipefail

# ============================================
# Arch WSL two-phase bootstrap (shim)
#  - Delegates to bin/arch-wsl
# ============================================

PHASE="${1:-phase1}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="${SCRIPT_DIR}"
CLI="${ROOT_DIR}/bin/arch-wsl"

if [[ ! -x "$CLI" ]]; then
  echo "[*] Making CLI executable: $CLI"
  chmod +x "$CLI" 2>/dev/null || true
fi

case "$PHASE" in
  phase1) exec "$CLI" phase1 ;;
  phase2) exec "$CLI" phase2 ;;
  *) echo "Usage: $0 [phase1|phase2]" >&2; exit 2 ;;
fi