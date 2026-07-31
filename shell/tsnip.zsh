# tsnip zsh widget — bind Ctrl+G to open the command palette.
# Load with:  eval "$(tsnip init zsh)"
# Or:         source <(tsnip init zsh)

tsnip-widget() {
  local selected
  selected="$(tsnip </dev/tty)" || return
  selected="${selected%"${selected##*[![:space:]]}"}"
  if [[ -n "$selected" ]]; then
    LBUFFER="${LBUFFER}${selected}"
    zle accept-line
  fi
}

zle -N tsnip-widget
bindkey '^G' tsnip-widget
