// Package search provides a UI-agnostic search engine over snippets.
// Designed to grow toward tags, aliases, and usage ranking without UI coupling.
package search

import (
	"sort"
	"strings"
	"unicode"

	"github.com/aashishvinu/tsnip/internal/models"
)

// Query describes a search request. Extra fields are reserved for future filters.
type Query struct {
	Text     string
	FolderID string // empty = search across all folders
	// Tags      []string
	// Favorites bool
}

// Result is a ranked match from the engine.
type Result struct {
	Ref       models.SnippetRef
	Score     int
	MatchedOn string
}

// Engine searches a corpus of snippets.
type Engine interface {
	Search(q Query, corpus []models.SnippetRef) []Result
}

// FuzzyEngine ranks snippets for command-palette filtering.
// Kept name for compatibility; matching is token/prefix oriented, not pure fuzzy.
type FuzzyEngine struct{}

// New returns the default search engine.
func New() Engine {
	return &FuzzyEngine{}
}

// NewFuzzyEngine is an alias for New.
func NewFuzzyEngine() Engine {
	return New()
}

// Search filters and ranks snippets.
// Empty text returns the corpus in order, optionally scoped to a folder.
// Non-empty text matches across the full corpus (folders ignored).
func (FuzzyEngine) Search(q Query, corpus []models.SnippetRef) []Result {
	text := strings.TrimSpace(q.Text)

	if text == "" {
		filtered := corpus
		if q.FolderID != "" {
			filtered = filterByFolder(corpus, q.FolderID)
		}
		out := make([]Result, len(filtered))
		for i, ref := range filtered {
			out[i] = Result{Ref: ref, Score: 0, MatchedOn: "order"}
		}
		return out
	}

	terms := splitTerms(text)
	if len(terms) == 0 {
		return nil
	}

	type scored struct {
		res   Result
		index int
	}
	var hits []scored
	for i, ref := range corpus {
		score, how := scoreRef(ref, terms)
		if score < 0 {
			continue
		}
		hits = append(hits, scored{
			res:   Result{Ref: ref, Score: score, MatchedOn: how},
			index: i,
		})
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].res.Score != hits[j].res.Score {
			return hits[i].res.Score > hits[j].res.Score
		}
		return hits[i].index < hits[j].index
	})

	out := make([]Result, len(hits))
	for i, h := range hits {
		out[i] = h.res
	}
	return out
}

func filterByFolder(corpus []models.SnippetRef, folderID string) []models.SnippetRef {
	out := make([]models.SnippetRef, 0, len(corpus))
	for _, ref := range corpus {
		if ref.FolderID == folderID {
			out = append(out, ref)
		}
	}
	return out
}

func splitTerms(text string) []string {
	fields := strings.Fields(strings.ToLower(normalize(text)))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return strings.Join(strings.Fields(s), " ")
}

func scoreRef(ref models.SnippetRef, terms []string) (int, string) {
	title := strings.ToLower(normalize(ref.Title))
	command := strings.ToLower(normalize(ref.Command))
	folder := strings.ToLower(normalize(ref.Folder))

	titleToks := tokenize(title)
	cmdToks := tokenize(command)
	folderToks := tokenize(folder)

	total := 0
	how := "command"
	for _, term := range terms {
		best := -1
		bestHow := ""

		if s := scoreField(command, cmdToks, term, 120); s > best {
			best, bestHow = s, "command"
		}
		if s := scoreField(title, titleToks, term, 100); s > best {
			best, bestHow = s, "title"
		}
		if s := scoreField(folder, folderToks, term, 70); s > best {
			best, bestHow = s, "folder"
		}

		if best < 0 {
			return -1, ""
		}
		total += best
		how = bestHow
	}
	return total, how
}

// scoreField ranks how well needle matches a field.
// Short needles (1–2 chars) only match whole tokens or token prefixes,
// so "ps" hits "docker ps" but not "pods".
func scoreField(hay string, tokens []string, needle string, weight int) int {
	if needle == "" || hay == "" {
		return -1
	}

	best := -1
	bump := func(v int) {
		if v > best {
			best = v
		}
	}

	if hay == needle {
		bump(weight + 80)
	}
	if strings.HasPrefix(hay, needle+" ") {
		bump(weight + 70)
	}

	for _, tok := range tokens {
		switch {
		case tok == needle:
			bump(weight + 60)
		case len(needle) == 1 && strings.HasPrefix(tok, needle):
			// Single letter: allow token prefixes ("g" → git).
			bump(weight + 20)
		case len(needle) >= 3 && strings.HasPrefix(tok, needle):
			bump(weight + 45)
		case len(needle) >= 3 && strings.Contains(tok, needle):
			bump(weight + 25)
		}
	}

	// Contiguous phrase / multi-char substring in the full field.
	if len(needle) >= 3 && strings.Contains(hay, needle) {
		bump(weight + 30)
	}

	if initialsMatch(tokens, needle) {
		bump(weight + 15)
	}

	// Loose subsequence only for longer patterns (command-palette "kgp" style).
	if len(needle) >= 3 && isSubsequence(hay, needle) {
		bump(weight / 5)
	}

	return best
}

func tokenize(s string) []string {
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		out = append(out, b.String())
		b.Reset()
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		flush()
	}
	flush()
	return out
}

func initialsMatch(tokens []string, query string) bool {
	if query == "" || len(tokens) == 0 {
		return false
	}
	var b strings.Builder
	for _, tok := range tokens {
		if tok == "" {
			continue
		}
		b.WriteByte(tok[0])
	}
	init := b.String()
	return strings.HasPrefix(init, query) || init == query
}

func isSubsequence(hay, needle string) bool {
	if needle == "" {
		return true
	}
	hr := []rune(hay)
	nr := []rune(needle)
	j := 0
	for i := 0; i < len(hr) && j < len(nr); i++ {
		if hr[i] == nr[j] {
			j++
		}
	}
	return j == len(nr)
}
