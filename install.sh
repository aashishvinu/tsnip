#!/usr/bin/env bash
# Install tsnip and wire Ctrl+G into your shell.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/aashishvinu/tsnip/main/install.sh | bash
#   ./install.sh          # from a local checkout

set -euo pipefail

BINDIR="${BINDIR:-$HOME/.local/bin}"
REPO="${REPO:-https://github.com/aashishvinu/tsnip.git}"
MODULE="github.com/aashishvinu/tsnip/cmd/tsnip@latest"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "tsnip: need '$1' on PATH" >&2
    exit 1
  }
}

install_binary() {
  mkdir -p "$BINDIR"
  export PATH="$BINDIR:$PATH"

  # Prefer building from this checkout when present.
  local here
  here="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" 2>/dev/null && pwd || true)"
  if [[ -n "${here:-}" && -f "$here/go.mod" && -d "$here/cmd/tsnip" ]]; then
    need go
    echo "→ building from local checkout"
    (cd "$here" && go build -o "$BINDIR/tsnip" ./cmd/tsnip)
    return
  fi

  if command -v go >/dev/null 2>&1; then
    echo "→ installing with go install"
    GOBIN="$BINDIR" go install "$MODULE"
    return
  fi

  need git
  need go
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  echo "→ cloning $REPO"
  git clone --depth 1 "$REPO" "$tmp/tsnip"
  (cd "$tmp/tsnip" && go build -o "$BINDIR/tsnip" ./cmd/tsnip)
}

shell_rc() {
  case "${SHELL##*/}" in
    zsh)  echo "$HOME/.zshrc" ;;
    bash) echo "$HOME/.bashrc" ;;
    *)    echo "" ;;
  esac
}

shell_kind() {
  case "${SHELL##*/}" in
    zsh)  echo zsh ;;
    bash) echo bash ;;
    *)    echo "" ;;
  esac
}

wire_shell() {
  local kind rc line
  kind="$(shell_kind)"
  rc="$(shell_rc)"
  if [[ -z "$kind" || -z "$rc" ]]; then
    echo "→ binary installed to $BINDIR/tsnip"
    echo "  add to your shell config:  eval \"\$(tsnip init zsh)\"  # or bash"
    return
  fi

  line="eval \"\$(tsnip init $kind)\""
  touch "$rc"
  if grep -Fqx "$line" "$rc" 2>/dev/null || grep -Fq 'tsnip init' "$rc" 2>/dev/null; then
    echo "→ shell hook already in $rc"
  else
    {
      echo ""
      echo "# tsnip — Ctrl+G command palette"
      echo "$line"
    } >>"$rc"
    echo "→ added Ctrl+G binding to $rc"
  fi
}

main() {
  install_binary
  if ! command -v tsnip >/dev/null 2>&1; then
    echo "→ note: add $BINDIR to your PATH"
  fi
  wire_shell
  echo ""
  echo "Done. Restart your shell (or: source $(shell_rc)) and press Ctrl+G."
}

main "$@"
