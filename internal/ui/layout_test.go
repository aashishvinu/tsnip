package ui_test

import (
	"testing"

	"github.com/aashishvinu/tsnip/internal/ui"
)

func TestHitTestPanels(t *testing.T) {
	l := ui.ComputeLayout(80, 24)
	if l.BodyH <= 0 || l.SearchY <= 0 {
		t.Fatalf("bad layout: %+v", l)
	}

	hit := l.HitTest(2, l.BodyTop)
	if hit.Panel != ui.PanelFolders || hit.Row != 0 {
		t.Fatalf("folder hit=%+v", hit)
	}

	hit = l.HitTest(l.FolderW+l.GapW+1, l.BodyTop+1)
	if hit.Panel != ui.PanelSnippets || hit.Row != 1 {
		t.Fatalf("snippet hit=%+v", hit)
	}

	hit = l.HitTest(10, l.SearchY)
	if hit.Panel != ui.PanelSearch {
		t.Fatalf("search hit=%+v", hit)
	}
}
