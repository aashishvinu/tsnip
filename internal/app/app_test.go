package app_test

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aashishvinu/tsnip/data"
	"github.com/aashishvinu/tsnip/internal/app"
	"github.com/aashishvinu/tsnip/internal/search"
	"github.com/aashishvinu/tsnip/internal/storage"
)

func newTestModel(t *testing.T) (app.Model, storage.Store) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "snippets.json")
	store := storage.NewJSONStore(path, data.Seed)
	doc, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	return app.New(app.Config{Data: doc, Store: store, Engine: search.New()}), store
}

func apply(m app.Model, msg tea.Msg) (app.Model, tea.Cmd) {
	next, cmd := m.Update(msg)
	return next.(app.Model), cmd
}

func goToFirstFolder(m app.Model) app.Model {
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyLeft})  // folders panel
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyDown})  // Recent → first folder
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyRight}) // snippets panel
	return m
}

func TestSelectPrintsCommand(t *testing.T) {
	m, _ := newTestModel(t)
	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = goToFirstFolder(m)
	m, cmd := apply(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if m.Cancelled() {
		t.Fatal("expected selection")
	}
	if m.SelectedCommand() == "" {
		t.Fatal("expected command")
	}
}

func TestSearchFilters(t *testing.T) {
	m, _ := newTestModel(t)
	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	// Typing a letter starts search without pressing /.
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("kub")})
	if m.View() == "" {
		t.Fatal("expected rendered view")
	}
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.SelectedCommand() == "" {
		t.Fatal("expected selected command from search")
	}
}

func TestCancel(t *testing.T) {
	m, _ := newTestModel(t)
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyCtrlQ})
	if !m.Cancelled() {
		t.Fatal("expected cancel")
	}
}

func TestSpaceCopies(t *testing.T) {
	m, _ := newTestModel(t)
	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = goToFirstFolder(m)
	m, cmd := apply(m, tea.KeyMsg{Type: tea.KeySpace})
	if cmd == nil {
		t.Fatal("copy should quit after copying")
	}
	if !m.Cancelled() {
		t.Fatal("copy should cancel without selecting for shell insert")
	}
	if m.SelectedCommand() != "" {
		t.Fatal("copy should not set selected command")
	}
}

func TestSpaceInSearchDoesNotCopy(t *testing.T) {
	m, _ := newTestModel(t)
	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("git")})
	m, cmd := apply(m, tea.KeyMsg{Type: tea.KeySpace})
	if cmd != nil {
		t.Fatal("space while searching should not quit")
	}
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("status")})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.SelectedCommand() == "" {
		t.Fatal("expected multi-word search to still select")
	}
}

func TestReorderFolderPersists(t *testing.T) {
	m, store := newTestModel(t)
	before, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	firstID := before.Folders[0].ID

	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyLeft})     // folders panel
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyDown})     // first real folder
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyCtrlDown})

	after, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if after.Folders[0].ID == firstID {
		t.Fatal("expected folder order to change and persist")
	}
	if after.Folders[1].ID != firstID {
		t.Fatalf("expected %s in second position, got %s", firstID, after.Folders[1].ID)
	}
}

func TestCtrlGQuits(t *testing.T) {
	m, _ := newTestModel(t)
	m, cmd := apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m, cmd = apply(m, tea.KeyMsg{Type: tea.KeyCtrlG})
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if !m.Cancelled() {
		t.Fatal("expected cancel on ctrl+g")
	}
}

func TestCreateFolder(t *testing.T) {
	m, store := newTestModel(t)
	before, _ := store.Load()
	n := len(before.Folders)

	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyCtrlF})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Custom")})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyEnter})

	after, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Folders) != n+1 {
		t.Fatalf("expected %d folders, got %d", n+1, len(after.Folders))
	}
	last := after.Folders[len(after.Folders)-1]
	if last.Name != "Custom" {
		t.Fatalf("name=%q", last.Name)
	}
}

func TestCreateSnippetMultiline(t *testing.T) {
	m, store := newTestModel(t)
	before, _ := store.Load()
	folderID := before.Folders[0].ID
	n := len(before.Folders[0].Snippets)

	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyCtrlN})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("echo one")})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("echo two")})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	after, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	got := after.Folders[0]
	if got.ID != folderID {
		t.Fatalf("expected folder %s", folderID)
	}
	if len(got.Snippets) != n+1 {
		t.Fatalf("expected %d snippets, got %d", n+1, len(got.Snippets))
	}
	cmd := got.Snippets[len(got.Snippets)-1].Command
	if cmd != "echo one\necho two" {
		t.Fatalf("command=%q", cmd)
	}
}

func TestCreateSnippetCancel(t *testing.T) {
	m, store := newTestModel(t)
	before, _ := store.Load()
	n := len(before.Folders[0].Snippets)

	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyCtrlN})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("should not save")})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyEscape})

	after, _ := store.Load()
	if len(after.Folders[0].Snippets) != n {
		t.Fatal("cancel should not persist snippet")
	}
}

func TestRecentTracksExecute(t *testing.T) {
	m, store := newTestModel(t)
	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = goToFirstFolder(m)
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyEnter})

	after, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Recent) == 0 {
		t.Fatal("expected recent entry after execute")
	}
}

func TestMouseClickSelectsFolder(t *testing.T) {
	m, _ := newTestModel(t)
	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})

	// Body starts at y=1; first folder under Recent is row 1 in the folder panel.
	m, _ = apply(m, tea.MouseMsg{
		X:      2,
		Y:      2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	m = goToFirstFolder(m) // ensure we can still navigate after mouse use
	m, cmd := apply(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || m.SelectedCommand() == "" {
		t.Fatal("expected selection after mouse + enter")
	}
}

func TestMouseLeftClickExecutes(t *testing.T) {
	m, _ := newTestModel(t)
	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = goToFirstFolder(m)

	// First snippet row at body top (y=1).
	m, cmd := apply(m, tea.MouseMsg{
		X:      30,
		Y:      1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	if cmd == nil {
		t.Fatal("expected quit on left-click")
	}
	if m.SelectedCommand() == "" {
		t.Fatal("expected command from left-click")
	}
}

func TestMouseRightClickCopies(t *testing.T) {
	m, _ := newTestModel(t)
	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = goToFirstFolder(m)

	m, cmd := apply(m, tea.MouseMsg{
		X:      30,
		Y:      1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonRight,
	})
	if cmd == nil {
		t.Fatal("expected quit on right-click copy")
	}
	if !m.Cancelled() {
		t.Fatal("copy should cancel without shell insert")
	}
	if m.SelectedCommand() != "" {
		t.Fatal("copy should not set selected command")
	}
}

func TestCtrlDDeletesSnippet(t *testing.T) {
	m, store := newTestModel(t)
	before, _ := store.Load()
	folderID := before.Folders[0].ID
	n := len(before.Folders[0].Snippets)
	firstID := before.Folders[0].Snippets[0].ID

	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = goToFirstFolder(m)
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	after, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	got := after.Folders[0]
	if got.ID != folderID {
		t.Fatalf("folder=%s", got.ID)
	}
	if len(got.Snippets) != n-1 {
		t.Fatalf("expected %d snippets, got %d", n-1, len(got.Snippets))
	}
	for _, s := range got.Snippets {
		if s.ID == firstID {
			t.Fatal("deleted snippet still present")
		}
	}
}

func TestCtrlDDeletesFolder(t *testing.T) {
	m, store := newTestModel(t)
	before, _ := store.Load()
	n := len(before.Folders)
	firstID := before.Folders[0].ID

	m, _ = apply(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyLeft})
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyDown}) // first real folder
	m, _ = apply(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	after, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Folders) != n-1 {
		t.Fatalf("expected %d folders, got %d", n-1, len(after.Folders))
	}
	for _, f := range after.Folders {
		if f.ID == firstID {
			t.Fatal("deleted folder still present")
		}
	}
}
