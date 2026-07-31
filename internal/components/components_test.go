package components_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/aashishvinu/tsnip/internal/components"
)

func TestMultiLineSnippetBlock(t *testing.T) {
	list := components.SnippetList{
		Items: []components.SnippetItem{
			{Code: "echo one"},
			{Code: "git add -A\ngit commit -m \"msg\""},
			{Code: "docker ps"},
		},
		Cursor:  1,
		Focused: true,
		Width:   40,
		Height:  10,
	}

	view := list.View()
	if lipgloss.Width(view) != 40 {
		t.Fatalf("width=%d", lipgloss.Width(view))
	}
	if lipgloss.Height(view) != 10 {
		t.Fatalf("height=%d", lipgloss.Height(view))
	}

	plain := view
	if !strings.Contains(plain, "git add -A") {
		t.Fatalf("missing first line of multi-line snippet:\n%s", plain)
	}
	if !strings.Contains(plain, "git commit") {
		t.Fatalf("missing second line of multi-line snippet:\n%s", plain)
	}
}

func TestSnippetShowsCodeNotTitle(t *testing.T) {
	list := components.SnippetList{
		Items: []components.SnippetItem{
			{Code: "kubectl logs -f {{pod}}"},
		},
		Cursor:  0,
		Focused: true,
		Width:   36,
		Height:  3,
	}
	view := list.View()
	if !strings.Contains(view, "kubectl logs") {
		t.Fatalf("expected command text in view:\n%s", view)
	}
}
