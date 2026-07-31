package search_test

import (
	"testing"

	"github.com/aashishvinu/tsnip/internal/models"
	"github.com/aashishvinu/tsnip/internal/search"
)

func sampleCorpus() []models.SnippetRef {
	return []models.SnippetRef{
		{FolderID: "kubernetes", SnippetID: "logs", Folder: "Kubernetes", Title: "Kubectl Logs", Command: "kubectl logs -f {{pod}}"},
		{FolderID: "kubernetes", SnippetID: "exec", Folder: "Kubernetes", Title: "Kubectl Exec", Command: "kubectl exec -it {{pod}} -- /bin/bash"},
		{FolderID: "kubernetes", SnippetID: "pods", Folder: "Kubernetes", Title: "Get Pods", Command: "kubectl get pods -n {{namespace}} -o wide"},
		{FolderID: "docker", SnippetID: "ps", Folder: "Docker", Title: "List Containers", Command: "docker ps -a"},
		{FolderID: "docker", SnippetID: "compose", Folder: "Docker", Title: "Compose Up", Command: "docker compose up -d --build\ndocker compose ps"},
		{FolderID: "git", SnippetID: "status", Folder: "Git", Title: "Status", Command: "git status -sb"},
		{FolderID: "git", SnippetID: "commit", Folder: "Git", Title: "Commit", Command: "git commit -m \"{{message}}\""},
	}
}

func ids(results []search.Result) []string {
	out := make([]string, len(results))
	for i, r := range results {
		out[i] = r.Ref.SnippetID
	}
	return out
}

func containsID(results []search.Result, id string) bool {
	for _, r := range results {
		if r.Ref.SnippetID == id {
			return true
		}
	}
	return false
}

func TestEmptyQueryRespectsFolder(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{FolderID: "kubernetes"}, sampleCorpus())
	if len(results) != 3 {
		t.Fatalf("expected 3 folder snippets, got %d", len(results))
	}
}

func TestPrefixFindsKubectl(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{Text: "kub"}, sampleCorpus())
	if len(results) == 0 {
		t.Fatal("expected matches for 'kub'")
	}
	for _, id := range []string{"logs", "exec", "pods"} {
		if !containsID(results, id) {
			t.Fatalf("expected %s in results: %v", id, ids(results))
		}
	}
}

func TestShortQueryDoesNotBleedIntoOtherTokens(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{Text: "ps"}, sampleCorpus())
	if !containsID(results, "ps") {
		t.Fatalf("expected docker ps match: %v", ids(results))
	}
	if !containsID(results, "compose") {
		t.Fatalf("expected compose (has ps on second line): %v", ids(results))
	}
	if containsID(results, "pods") {
		t.Fatalf("'ps' should not match 'pods': %v", ids(results))
	}
}

func TestMultiWordAND(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{Text: "git status"}, sampleCorpus())
	if len(results) != 1 || results[0].Ref.SnippetID != "status" {
		t.Fatalf("expected only git status, got %v", ids(results))
	}
}

func TestFindsAcrossFolders(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{Text: "status"}, sampleCorpus())
	if !containsID(results, "status") {
		t.Fatalf("expected git status in results: %v", ids(results))
	}
}

func TestFolderNameMatch(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{Text: "docker"}, sampleCorpus())
	if !containsID(results, "ps") || !containsID(results, "compose") {
		t.Fatalf("expected docker folder snippets: %v", ids(results))
	}
}

func TestInitialsMatch(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{Text: "kgp"}, sampleCorpus())
	if !containsID(results, "pods") {
		t.Fatalf("expected 'kgp' to match kubectl get pods: %v", ids(results))
	}
}

func TestMultilineCommandSearchable(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{Text: "compose"}, sampleCorpus())
	if !containsID(results, "compose") {
		t.Fatalf("expected multiline compose snippet: %v", ids(results))
	}
}

func TestRanksExactTokenAboveLoose(t *testing.T) {
	engine := search.New()
	results := engine.Search(search.Query{Text: "ps"}, sampleCorpus())
	if len(results) == 0 || results[0].Ref.SnippetID != "ps" {
		t.Fatalf("expected docker ps ranked first, got %v", ids(results))
	}
}
