# tsnip bash widget — bind Ctrl+G to open the command palette.
# Load with:  eval "$(tsnip init bash)"
# Or:         source <(tsnip init bash)

tsnip-widget() {
  local selected
  selected="$(tsnip </dev/tty)" || return
  # Trim trailing whitespace/newlines from tsnip stdout.
  selected="${selected%"${selected##*[![:space:]]}"}"
  if [[ -n "$selected" ]]; then
    READLINE_LINE="${READLINE_LINE:0:$READLINE_POINT}${selected}${READLINE_LINE:$READLINE_POINT}"
    READLINE_POINT=$((READLINE_POINT + ${#selected}))
    # bind -x cannot accept the line; run it directly like Enter would.
    history -s "$READLINE_LINE"
    eval "$READLINE_LINE"
    READLINE_LINE=""
    READLINE_POINT=0
  fi
}

bind -x '"\C-g": tsnip-widget'
