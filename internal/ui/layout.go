package ui

// Layout describes the on-screen regions for mouse hit-testing.
// Must stay in sync with View().
type Layout struct {
	Width  int
	Height int

	TopGap    int
	BottomGap int
	SepH      int
	SearchH   int

	BodyTop int
	BodyH   int

	FolderW  int
	GapW     int
	SnippetW int

	SearchY int
}

// Panel identifies which chrome region was hit.
type Panel int

const (
	PanelNone Panel = iota
	PanelFolders
	PanelSnippets
	PanelSearch
)

// Hit is the result of mapping a mouse coordinate into the UI.
type Hit struct {
	Panel Panel
	Row   int // 0-based row within the panel (folders/snippets body)
}

// ComputeLayout returns geometry for the current terminal size.
func ComputeLayout(width, height int) Layout {
	l := Layout{
		Width:     width,
		Height:    height,
		TopGap:    1,
		BottomGap: 1,
		SepH:      1,
		SearchH:   1,
		GapW:      2,
	}
	if width <= 0 || height <= 0 {
		return l
	}

	chrome := l.TopGap + l.BottomGap + l.SepH + l.SearchH
	l.BodyH = height - chrome
	if l.BodyH < 1 {
		l.BodyH = 1
	}
	l.BodyTop = l.TopGap

	l.FolderW = FolderPanelWidth
	if width < 56 {
		l.FolderW = max(10, width/3)
	}
	l.SnippetW = width - l.FolderW - l.GapW
	if l.SnippetW < 8 {
		l.SnippetW = 8
		l.FolderW = max(8, width-l.SnippetW-l.GapW)
	}

	l.SearchY = l.BodyTop + l.BodyH + l.BottomGap + l.SepH
	return l
}

// HitTest maps absolute terminal coordinates to a UI panel.
func (l Layout) HitTest(x, y int) Hit {
	if l.Width <= 0 || l.Height <= 0 || x < 0 || y < 0 || x >= l.Width || y >= l.Height {
		return Hit{Panel: PanelNone, Row: -1}
	}

	if y == l.SearchY {
		return Hit{Panel: PanelSearch, Row: 0}
	}

	if y < l.BodyTop || y >= l.BodyTop+l.BodyH {
		return Hit{Panel: PanelNone, Row: -1}
	}

	row := y - l.BodyTop
	switch {
	case x < l.FolderW:
		return Hit{Panel: PanelFolders, Row: row}
	case x >= l.FolderW+l.GapW:
		return Hit{Panel: PanelSnippets, Row: row}
	default:
		return Hit{Panel: PanelNone, Row: -1}
	}
}
