// Package keymap defines keyboard bindings for tsnip.
package keymap

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds all interactive bindings.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	NextPanel   key.Binding
	PrevPanel   key.Binding
	Enter       key.Binding
	Escape      key.Binding
	Quit        key.Binding
	Copy        key.Binding
	Search      key.Binding
	MoveUp      key.Binding
	MoveDown    key.Binding
	ClearSearch key.Binding
	NewSnippet  key.Binding
	NewFolder   key.Binding
	Delete      key.Binding
	Save        key.Binding
}

// Default returns the standard key map.
func Default() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "folders"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "snippets"),
		),
		NextPanel: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		PrevPanel: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "execute"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear / quit"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+g", "ctrl+q", "ctrl+z"),
			key.WithHelp("ctrl+g", "quit"),
		),
		Copy: key.NewBinding(
			key.WithKeys(" ", "space"),
			key.WithHelp("space", "copy"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("ctrl+up", "alt+up"),
			key.WithHelp("ctrl+↑", "move up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("ctrl+down", "alt+down"),
			key.WithHelp("ctrl+↓", "move down"),
		),
		ClearSearch: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "clear search"),
		),
		NewSnippet: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "new snippet"),
		),
		NewFolder: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "new folder"),
		),
		Delete: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "delete"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
	}
}

// CreateFolderHelp is shown while naming a folder.
func (k KeyMap) CreateFolderHelp() string {
	return "enter save  ·  esc cancel"
}

// CreateSnippetHelp is shown while composing a snippet.
func (k KeyMap) CreateSnippetHelp() string {
	return "enter newline  ·  ctrl+s save  ·  esc cancel"
}
