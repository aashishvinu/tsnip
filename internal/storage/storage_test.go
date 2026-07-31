package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aashishvinu/tsnip/internal/models"
	"github.com/aashishvinu/tsnip/internal/storage"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snippets.json")
	store := storage.NewJSONStore(path, nil)

	doc := &models.Data{
		Folders: []models.Folder{
			{
				ID:   "git",
				Name: "Git",
				Snippets: []models.Snippet{
					{ID: "st", Title: "Status", Command: "git status"},
				},
			},
		},
	}
	if err := store.Save(doc); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Folders) != 1 || got.Folders[0].Snippets[0].Command != "git status" {
		t.Fatalf("unexpected doc: %+v", got)
	}
}

func TestEnsureSeedsOnce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snippets.json")
	seed := []byte(`{"folders":[{"id":"a","name":"A","snippets":[]}]}`)
	store := storage.NewJSONStore(path, seed)

	doc, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Folders) != 1 {
		t.Fatalf("expected seed folder")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	doc.Folders[0].Name = "Changed"
	if err := store.Save(doc); err != nil {
		t.Fatal(err)
	}
	again, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if again.Folders[0].Name != "Changed" {
		t.Fatal("Load should not overwrite existing data")
	}
}

func TestMoveFolderAndSnippet(t *testing.T) {
	doc := &models.Data{
		Folders: []models.Folder{
			{ID: "a", Name: "A", Snippets: []models.Snippet{{ID: "1", Title: "One", Command: "one"}, {ID: "2", Title: "Two", Command: "two"}}},
			{ID: "b", Name: "B", Snippets: nil},
		},
	}
	if !doc.MoveFolder("a", 1) {
		t.Fatal("move folder failed")
	}
	if doc.Folders[0].ID != "b" {
		t.Fatalf("expected b first, got %s", doc.Folders[0].ID)
	}
	doc.MoveFolder("a", -1)
	if !doc.MoveSnippet("a", "1", 1) {
		t.Fatal("move snippet failed")
	}
	if doc.Folders[0].Snippets[0].ID != "2" {
		t.Fatalf("expected snippet 2 first")
	}
}
