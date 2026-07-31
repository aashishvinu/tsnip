// Package ui composes the full-screen layout from smaller components.
// Layout: top gap, body panels, bottom gap, separator, search.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aashishvinu/tsnip/internal/components"
	"github.com/aashishvinu/tsnip/internal/theme"
)

// FolderPanelWidth is the preferred width of the left folder column.
const FolderPanelWidth = 20

// Model holds layout inputs for a single frame.
type Model struct {
	Width  int
	Height int

	Folders   components.FolderList
	Snippets  components.SnippetList
	Search    components.SearchBar
	Searching bool

	Composing    bool
	ComposeCode  string
	ComposeLabel string
}

// View renders the complete tsnip screen with exact terminal dimensions.
func (m Model) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return ""
	}

	const (
		topGap    = 1
		bottomGap = 1
		sepH      = 1
		searchH   = 1
	)

	chrome := topGap + bottomGap + sepH + searchH
	bodyH := m.Height - chrome
	if bodyH < 1 {
		bodyH = 1
	}

	folderW := FolderPanelWidth
	if m.Width < 56 {
		folderW = max(10, m.Width/3)
	}
	gapW := 2
	snippetW := m.Width - folderW - gapW
	if snippetW < 8 {
		snippetW = 8
		folderW = max(8, m.Width-snippetW-gapW)
	}

	m.Folders.Width = folderW
	m.Folders.Height = bodyH
	m.Snippets.Width = snippetW
	m.Snippets.Height = bodyH
	m.Search.Width = m.Width

	if m.Composing {
		if m.ComposeCode == "" {
			label := m.ComposeLabel
			if label == "" {
				label = "start typing…"
			}
			m.Snippets = components.SnippetList{
				Heading: m.Snippets.Heading,
				Empty:   label,
				Width:   snippetW,
				Height:  bodyH,
				Focused: true,
			}
		} else {
			m.Snippets = components.SnippetList{
				Heading: m.Snippets.Heading,
				Items:   []components.SnippetItem{{Code: m.ComposeCode}},
				Cursor:  0,
				Focused: true,
				Width:   snippetW,
				Height:  bodyH,
			}
		}
	}

	left := m.Folders.View()
	right := m.Snippets.View()
	gapLines := make([]string, bodyH)
	for i := range gapLines {
		gapLines[i] = theme.NormalRow.Width(gapW).MaxWidth(gapW).Render(strings.Repeat(" ", gapW))
	}
	gap := strings.Join(gapLines, "\n")

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
	body = lipgloss.NewStyle().
		Width(m.Width).
		Height(bodyH).
		MaxHeight(bodyH).
		MaxWidth(m.Width).
		Background(theme.Default.Bg).
		Render(body)

	blank := blankRow(m.Width)
	sep := separatorRow(m.Width)
	search := m.Search.View()

	frame := lipgloss.JoinVertical(lipgloss.Left, blank, body, blank, sep, search)
	return lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		MaxWidth(m.Width).
		MaxHeight(m.Height).
		Background(theme.Default.Bg).
		Render(frame)
}

func blankRow(width int) string {
	return theme.NormalRow.Width(width).MaxWidth(width).Render(strings.Repeat(" ", width))
}

func separatorRow(width int) string {
	if width <= 0 {
		return ""
	}
	line := strings.Repeat("─", width)
	return theme.Separator.Width(width).MaxWidth(width).Render(line)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
