package models_test

import (
	"testing"

	"github.com/aashishvinu/tsnip/internal/models"
)

func TestAddFolderAndSnippet(t *testing.T) {
	d := &models.Data{}
	fi, err := d.AddFolder("My Tools")
	if err != nil {
		t.Fatal(err)
	}
	if fi != 0 || d.Folders[0].ID != "my-tools" {
		t.Fatalf("folder: %+v", d.Folders[0])
	}

	si, err := d.AddSnippet("my-tools", "", "echo hello\necho world")
	if err != nil {
		t.Fatal(err)
	}
	if si != 0 {
		t.Fatalf("snippet idx %d", si)
	}
	s := d.Folders[0].Snippets[0]
	if s.Command != "echo hello\necho world" {
		t.Fatalf("command=%q", s.Command)
	}
	if s.Title != "echo hello" {
		t.Fatalf("title=%q", s.Title)
	}
}

func TestAddFolderRejectsEmpty(t *testing.T) {
	d := &models.Data{}
	if _, err := d.AddFolder("  "); err == nil {
		t.Fatal("expected error")
	}
}

func TestUniqueIDs(t *testing.T) {
	d := &models.Data{}
	_, _ = d.AddFolder("Git")
	_, _ = d.AddFolder("Git")
	if d.Folders[1].ID != "git-2" {
		t.Fatalf("got id %q", d.Folders[1].ID)
	}
}

func TestTouchRecent(t *testing.T) {
	d := &models.Data{
		Folders: []models.Folder{
			{ID: "git", Name: "Git", Snippets: []models.Snippet{
				{ID: "a", Title: "A", Command: "echo a"},
				{ID: "b", Title: "B", Command: "echo b"},
			}},
		},
	}
	d.TouchRecent("git", "b")
	d.TouchRecent("git", "a")
	d.TouchRecent("git", "b")
	got := d.RecentSnippets()
	if len(got) != 2 || got[0].SnippetID != "b" || got[1].SnippetID != "a" {
		t.Fatalf("recent=%+v", got)
	}
}

func TestDeleteSnippetAndFolder(t *testing.T) {
	d := &models.Data{
		Folders: []models.Folder{
			{ID: "git", Name: "Git", Snippets: []models.Snippet{
				{ID: "a", Title: "A", Command: "echo a"},
				{ID: "b", Title: "B", Command: "echo b"},
			}},
			{ID: "docker", Name: "Docker", Snippets: []models.Snippet{
				{ID: "ps", Title: "PS", Command: "docker ps"},
			}},
		},
	}
	d.TouchRecent("git", "a")
	d.TouchRecent("docker", "ps")

	if !d.DeleteSnippet("git", "a") {
		t.Fatal("expected snippet delete")
	}
	if len(d.Folders[0].Snippets) != 1 || d.Folders[0].Snippets[0].ID != "b" {
		t.Fatalf("snippets=%+v", d.Folders[0].Snippets)
	}
	if len(d.Recent) != 1 || d.Recent[0].SnippetID != "ps" {
		t.Fatalf("recent after snippet delete=%+v", d.Recent)
	}

	if !d.DeleteFolder("docker") {
		t.Fatal("expected folder delete")
	}
	if len(d.Folders) != 1 || d.Folders[0].ID != "git" {
		t.Fatalf("folders=%+v", d.Folders)
	}
	if len(d.Recent) != 0 {
		t.Fatalf("recent should be empty, got %+v", d.Recent)
	}
	if d.DeleteFolder(models.RecentFolderID) {
		t.Fatal("must not delete virtual Recent")
	}
}
