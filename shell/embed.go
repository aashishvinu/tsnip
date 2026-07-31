// Package shell embeds shell integration snippets for Ctrl+G.
package shell

import _ "embed"

//go:embed tsnip.zsh
var Zsh string

//go:embed tsnip.bash
var Bash string
