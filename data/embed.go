// Package data embeds the default seed snippets shipped with tsnip.
package data

import _ "embed"

// Seed is the default snippets.json written on first run.
//
//go:embed seed.json
var Seed []byte
