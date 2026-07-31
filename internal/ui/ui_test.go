package ui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/aashishvinu/tsnip/internal/components"
	"github.com/aashishvinu/tsnip/internal/ui"
)

func TestViewExactDimensions(t *testing.T) {
	cases := []struct{ w, h int }{
		{80, 24},
		{120, 40},
		{60, 20},
		{40, 15},
	}

	for _, tc := range cases {
		m := ui.Model{
			Width:  tc.w,
			Height: tc.h,
			Folders: components.FolderList{
				Items: []components.FolderItem{
					{Name: "Recent", Special: true},
					{Name: "Kubernetes"},
					{Name: "Docker"},
				},
				Cursor:  0,
				Focused: true,
			},
			Snippets: components.SnippetList{
				Items: []components.SnippetItem{
					{Code: "kubectl logs -f {{pod}}"},
					{Code: "kubectl exec -it {{pod}} -- /bin/bash"},
					{Code: "cat <<'EOF' | kubectl apply -f -\napiVersion: v1\nkind: ConfigMap\nEOF"},
				},
				Cursor:  2,
				Focused: false,
			},
			Search: components.SearchBar{
				Query:  "kub",
				Cursor: true,
				Active: true,
				Right:  "3 matches",
			},
		}

		view := m.View()
		if view == "" {
			t.Fatalf("%dx%d: empty view", tc.w, tc.h)
		}

		gotH := lipgloss.Height(view)
		gotW := lipgloss.Width(view)
		if gotH != tc.h {
			t.Errorf("%dx%d: height=%d want %d", tc.w, tc.h, gotH, tc.h)
		}
		if gotW != tc.w {
			t.Errorf("%dx%d: width=%d want %d", tc.w, tc.h, gotW, tc.w)
		}

		for i, line := range strings.Split(view, "\n") {
			if lw := lipgloss.Width(line); lw != tc.w {
				t.Errorf("%dx%d: line %d width=%d want %d\n%s", tc.w, tc.h, i, lw, tc.w, line)
				break
			}
		}
	}
}

func TestFolderListExactSize(t *testing.T) {
	f := components.FolderList{
		Items: []components.FolderItem{
			{Name: "Recent", Special: true},
			{Name: "A"},
			{Name: "B"},
		},
		Cursor:  1,
		Focused: true,
		Width:   18,
		Height:  6,
	}
	view := f.View()
	if lipgloss.Width(view) != 18 {
		t.Fatalf("width=%d", lipgloss.Width(view))
	}
	if lipgloss.Height(view) != 6 {
		t.Fatalf("height=%d", lipgloss.Height(view))
	}
}
