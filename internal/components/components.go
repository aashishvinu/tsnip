// Package components provides reusable TUI widgets for tsnip.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aashishvinu/tsnip/internal/theme"
)

const rowPad = 2

// FolderItem is one row in the folder panel.
type FolderItem struct {
	Name    string
	Special bool
}

// FolderList renders the left-hand folder panel.
type FolderList struct {
	Items    []FolderItem
	Cursor   int
	Focused  bool
	Width    int
	Height   int
	Disabled bool
}

// View renders the folder list to exactly Width x Height cells.
func (f FolderList) View() string {
	if f.Width <= 0 || f.Height <= 0 {
		return ""
	}

	lines := make([]string, 0, f.Height)
	start := scrollOffset(f.Cursor, len(f.Items), f.Height)
	end := min(start+f.Height, len(f.Items))

	for i := start; i < end; i++ {
		item := f.Items[i]
		lines = append(lines, paintRow(
			item.Name, "", item.Special,
			i == f.Cursor, f.Focused && !f.Disabled, f.Width,
		))
	}
	for len(lines) < f.Height {
		lines = append(lines, paintBlank(f.Width))
	}
	return strings.Join(lines, "\n")
}

// IndexAt maps a viewport row to a folder item index, or -1 if none.
func (f FolderList) IndexAt(row int) int {
	if row < 0 || row >= f.Height || len(f.Items) == 0 {
		return -1
	}
	start := scrollOffset(f.Cursor, len(f.Items), f.Height)
	idx := start + row
	if idx < 0 || idx >= len(f.Items) {
		return -1
	}
	return idx
}

// SnippetItem is one snippet in the panel. Code may contain multiple lines.
type SnippetItem struct {
	Code string
	Meta string
}

// SnippetList renders snippet commands (including multi-line blocks).
type SnippetList struct {
	Items   []SnippetItem
	Cursor  int
	Focused bool
	Width   int
	Height  int
	Empty   string
	Heading string
}

// View renders the snippet list to exactly Width x Height cells.
func (s SnippetList) View() string {
	if s.Width <= 0 || s.Height <= 0 {
		return ""
	}

	out := make([]string, 0, s.Height)
	listH := s.Height
	if s.Heading != "" && listH > 0 {
		out = append(out, paintHeading(s.Heading, s.Width))
		listH--
	}

	if listH <= 0 {
		return strings.Join(out, "\n")
	}

	if len(s.Items) == 0 {
		msg := s.Empty
		if msg == "" {
			msg = "no snippets"
		}
		line := theme.EmptyState.Width(s.Width).MaxWidth(s.Width).Render(
			padVisible(strings.Repeat(" ", rowPad)+fit(msg, max(0, s.Width-rowPad)), s.Width),
		)
		out = append(out, line)
		for len(out) < s.Height {
			out = append(out, paintBlank(s.Width))
		}
		return strings.Join(out, "\n")
	}

	blocks := make([][]string, len(s.Items))
	heights := make([]int, len(s.Items))
	for i, item := range s.Items {
		blocks[i] = s.renderBlock(item, i == s.Cursor)
		heights[i] = len(blocks[i])
		if i < len(s.Items)-1 {
			heights[i]++
		}
	}

	start := blockScrollOffset(s.Cursor, heights, listH)
	for i := start; i < len(s.Items) && len(out) < s.Height; i++ {
		for _, line := range blocks[i] {
			if len(out) >= s.Height {
				break
			}
			out = append(out, line)
		}
		if i < len(s.Items)-1 && len(out) < s.Height {
			out = append(out, paintBlank(s.Width))
		}
	}
	for len(out) < s.Height {
		out = append(out, paintBlank(s.Width))
	}
	return strings.Join(out, "\n")
}

func (s SnippetList) renderBlock(item SnippetItem, selected bool) []string {
	code := strings.ReplaceAll(item.Code, "\r\n", "\n")
	code = strings.ReplaceAll(code, "\r", "\n")
	rawLines := strings.Split(code, "\n")
	if len(rawLines) == 0 {
		rawLines = []string{""}
	}
	if len(rawLines) > 1 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	out := make([]string, 0, len(rawLines))
	for i, text := range rawLines {
		meta := ""
		if i == 0 {
			meta = item.Meta
		}
		out = append(out, paintRow(text, meta, false, selected, s.Focused, s.Width))
	}
	return out
}

// IndexAt maps a viewport row to a snippet item index, or -1 if none
// (heading, spacer, or empty area).
func (s SnippetList) IndexAt(row int) int {
	if row < 0 || row >= s.Height || len(s.Items) == 0 {
		return -1
	}

	listH := s.Height
	if s.Heading != "" {
		if row == 0 {
			return -1
		}
		row--
		listH--
	}
	if listH <= 0 || row < 0 || row >= listH {
		return -1
	}

	heights := make([]int, len(s.Items))
	for i, item := range s.Items {
		heights[i] = len(s.renderBlock(item, false))
		if i < len(s.Items)-1 {
			heights[i]++
		}
	}

	start := blockScrollOffset(s.Cursor, heights, listH)
	y := 0
	for i := start; i < len(s.Items); i++ {
		h := heights[i]
		if row >= y && row < y+h {
			// Spacer line between blocks maps to the previous item only for
			// the content rows; treat spacer as belonging to this item's gap
			// after its content — still select this item if within block.
			blockH := len(s.renderBlock(s.Items[i], false))
			if row < y+blockH {
				return i
			}
			return -1
		}
		y += h
		if y > row {
			break
		}
	}
	return -1
}
type SearchBar struct {
	Prompt      string
	Placeholder string
	Query       string
	Width       int
	Cursor      bool
	Active      bool
	Right       string // context, match count, or status on the trailing edge
}

// View renders the search content line to exactly Width cells.
func (s SearchBar) View() string {
	if s.Width <= 0 {
		return ""
	}

	label := s.Prompt
	if label == "" {
		label = "/"
	}
	placeholder := s.Placeholder
	if placeholder == "" {
		placeholder = "search snippets"
	}

	promptStyle := theme.SearchPrompt
	if !s.Active && s.Query == "" {
		promptStyle = theme.Faint
	}
	prompt := promptStyle.Render(label)

	right := ""
	rightW := 0
	if s.Right != "" {
		right = theme.Faint.Render(fit(s.Right, max(8, s.Width/3)))
		rightW = lipgloss.Width(right) + 2 // gap before right
	}

	// Inset one cell on each side to align with list row padding.
	inner := max(1, s.Width-2)
	rest := inner - lipgloss.Width(prompt) - 1 - rightW
	if rest < 1 {
		line := padVisible(" "+prompt, s.Width)
		return theme.NormalRow.Width(s.Width).MaxWidth(s.Width).Render(line)
	}

	display := s.Query
	if i := strings.LastIndexByte(display, '\n'); i >= 0 {
		display = display[i+1:]
	}

	var input string
	switch {
	case display == "" && s.Query == "":
		if s.Cursor {
			input = theme.Subtle.Render(fit(placeholder, max(0, rest-1))) + theme.SearchCursor.Render(" ")
		} else {
			input = theme.Subtle.Render(fit(placeholder, rest))
		}
	case s.Cursor:
		input = theme.SearchText.Render(fit(display, max(0, rest-1))) + theme.SearchCursor.Render(" ")
	default:
		input = theme.SearchText.Render(fit(display, rest))
	}

	left := padVisible(prompt+" "+input, rest+lipgloss.Width(prompt)+1)
	content := left
	if right != "" {
		gap := inner - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 1 {
			gap = 1
		}
		content = left + strings.Repeat(" ", gap) + right
	}
	line := " " + padVisible(content, inner) + " "
	return theme.NormalRow.Width(s.Width).MaxWidth(s.Width).Render(padVisible(line, s.Width))
}

func paintHeading(text string, width int) string {
	if width <= 0 {
		return ""
	}
	label := strings.Repeat(" ", rowPad) + fit(text, max(0, width-rowPad))
	return theme.PanelHeader.Width(width).MaxWidth(width).Render(padVisible(label, width))
}

// paintRow draws a full-width rectangular row. One style covers the entire cell
// so selection is a clean solid bar (no broken grey patches).
func paintRow(text, meta string, special, selected, focused bool, width int) string {
	if width <= 0 {
		return ""
	}

	style := theme.NormalRow
	metaStyle := theme.Faint
	switch {
	case selected && focused:
		style = theme.SelectedRow
		metaStyle = theme.SelectedMeta
	case selected && !focused:
		style = theme.SelectedRowDim
		metaStyle = theme.SelectedMetaDim
	case special:
		style = theme.SpecialRow
	}

	inner := width - rowPad
	if inner < 1 {
		return style.Width(width).MaxWidth(width).Render(strings.Repeat(" ", width))
	}

	pad := strings.Repeat(" ", rowPad)
	content := pad + fit(text, inner)

	if meta != "" {
		metaBudget := min(18, max(6, inner/3))
		metaText := fit(meta, metaBudget)
		metaW := lipgloss.Width(metaText)
		titleW := inner - metaW - 2
		if titleW >= 4 {
			styledMeta := metaStyle.Render(metaText)
			content = pad + padVisible(fit(text, titleW), titleW) + "  " + styledMeta
		}
	}

	return style.Width(width).MaxWidth(width).Render(padVisible(content, width))
}

func paintBlank(width int) string {
	if width <= 0 {
		return ""
	}
	return theme.NormalRow.Width(width).MaxWidth(width).Render(strings.Repeat(" ", width))
}

func blockScrollOffset(cursor int, heights []int, viewport int) int {
	if cursor < 0 || cursor >= len(heights) || viewport <= 0 {
		return 0
	}
	selH := heights[cursor]
	if selH > viewport {
		return cursor
	}
	start := 0
	for {
		top := 0
		for i := start; i < cursor; i++ {
			top += heights[i]
		}
		if top+selH <= viewport {
			return start
		}
		if start >= cursor {
			return cursor
		}
		start++
	}
}

func scrollOffset(cursor, total, visible int) int {
	if total <= visible || cursor < visible {
		return 0
	}
	return cursor - visible + 1
}

func fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	if width == 1 {
		return "…"
	}
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func padVisible(s string, width int) string {
	w := lipgloss.Width(s)
	switch {
	case w == width:
		return s
	case w > width:
		return fit(s, width)
	default:
		return s + strings.Repeat(" ", width-w)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
