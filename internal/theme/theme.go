// Package theme centralizes visual styling for tsnip.
package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Palette — quiet charcoal with a single soft accent.
type Palette struct {
	Bg          lipgloss.Color
	Surface     lipgloss.Color
	Border      lipgloss.Color
	Text        lipgloss.Color
	TextMuted   lipgloss.Color
	TextFaint   lipgloss.Color
	Accent      lipgloss.Color
	SelectionBg lipgloss.Color
	SelectionFg lipgloss.Color
	StatusBg    lipgloss.Color
	StatusFg    lipgloss.Color
}

// Default is the primary theme.
var Default = Palette{
	Bg:          lipgloss.Color("#0e0e0e"),
	Surface:     lipgloss.Color("#181818"),
	Border:      lipgloss.Color("#2b2b2b"),
	Text:        lipgloss.Color("#d0d0d0"),
	TextMuted:   lipgloss.Color("#8e8e8e"),
	TextFaint:   lipgloss.Color("#5a5a5a"),
	Accent:      lipgloss.Color("#7aa2f7"),
	SelectionBg: lipgloss.Color("#2c2c2c"),
	SelectionFg: lipgloss.Color("#f2f2f2"),
	StatusBg:    lipgloss.Color("#181818"),
	StatusFg:    lipgloss.Color("#8e8e8e"),
}

// Styles — components must use these (no nested conflicting backgrounds).
var (
	App             lipgloss.Style
	Subtle          lipgloss.Style
	Faint           lipgloss.Style
	NormalRow       lipgloss.Style
	SelectedRow     lipgloss.Style
	SelectedRowDim  lipgloss.Style
	SelectedMeta    lipgloss.Style
	SelectedMetaDim lipgloss.Style
	SpecialRow      lipgloss.Style
	SearchPrompt    lipgloss.Style
	SearchText      lipgloss.Style
	SearchCursor    lipgloss.Style
	Footer          lipgloss.Style
	StatusBar       lipgloss.Style
	EmptyState      lipgloss.Style
	Help            lipgloss.Style
	PanelHeader     lipgloss.Style
	Separator       lipgloss.Style
)

func init() {
	// Shell widgets capture stdout (command substitution), so detect color
	// capabilities from stderr — the writer Bubble Tea actually paints to.
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))

	p := Default

	App = lipgloss.NewStyle().
		Background(p.Bg).
		Foreground(p.Text)

	Subtle = lipgloss.NewStyle().
		Foreground(p.TextMuted)

	Faint = lipgloss.NewStyle().
		Foreground(p.TextFaint)

	// Every row paints a full-width background so selection is a clean bar.
	NormalRow = lipgloss.NewStyle().
		Foreground(p.Text).
		Background(p.Bg)

	SelectedRow = lipgloss.NewStyle().
		Foreground(p.SelectionFg).
		Background(p.SelectionBg).
		Bold(true)

	SelectedRowDim = lipgloss.NewStyle().
		Foreground(p.TextMuted).
		Background(p.Surface)

	SelectedMeta = lipgloss.NewStyle().
		Foreground(p.TextMuted).
		Background(p.SelectionBg)

	SelectedMetaDim = lipgloss.NewStyle().
		Foreground(p.TextFaint).
		Background(p.Surface)

	SpecialRow = lipgloss.NewStyle().
		Foreground(p.Accent).
		Background(p.Bg)

	SearchPrompt = lipgloss.NewStyle().
		Foreground(p.Accent).
		Bold(true)

	SearchText = lipgloss.NewStyle().
		Foreground(p.Text)

	SearchCursor = lipgloss.NewStyle().
		Foreground(p.Bg).
		Background(p.Accent)

	Footer = lipgloss.NewStyle().
		Foreground(p.Text).
		Background(p.Bg).
		Padding(0, 1)

	StatusBar = lipgloss.NewStyle().
		Foreground(p.StatusFg).
		Background(p.Bg).
		Padding(0, 1)

	EmptyState = lipgloss.NewStyle().
		Foreground(p.TextFaint).
		Background(p.Bg)

	Help = lipgloss.NewStyle().
		Foreground(p.TextFaint)

	PanelHeader = lipgloss.NewStyle().
		Foreground(p.TextMuted).
		Background(p.Bg).
		Bold(true)

	Separator = lipgloss.NewStyle().
		Foreground(p.Border).
		Background(p.Bg)
}
