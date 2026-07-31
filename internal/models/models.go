// Package models defines the domain types for folders and snippets.
package models

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	// RecentFolderID is the virtual folder id for recently used snippets.
	RecentFolderID = "__recent__"
	// RecentLimit is the maximum number of recent entries kept.
	RecentLimit = 10
)

// Snippet is a reusable command entry.
type Snippet struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Command string `json:"command"`
}

// Folder groups related snippets. Slice order is display order.
type Folder struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Snippets []Snippet `json:"snippets"`
}

// RecentEntry points at a snippet that was recently executed or copied.
type RecentEntry struct {
	FolderID  string `json:"folder_id"`
	SnippetID string `json:"snippet_id"`
}

// Data is the root persisted document.
type Data struct {
	Folders []Folder      `json:"folders"`
	Recent  []RecentEntry `json:"recent,omitempty"`
}

// SnippetRef uniquely identifies a snippet for search results and selection.
type SnippetRef struct {
	FolderID  string
	SnippetID string
	Folder    string
	Title     string
	Command   string
}

// FindFolder returns the folder with the given id and its index, or (-1, nil).
func (d *Data) FindFolder(id string) (int, *Folder) {
	if d == nil {
		return -1, nil
	}
	for i := range d.Folders {
		if d.Folders[i].ID == id {
			return i, &d.Folders[i]
		}
	}
	return -1, nil
}

// FindSnippet returns folder index, snippet index, and the snippet for the given ids.
func (d *Data) FindSnippet(folderID, snippetID string) (folderIdx, snippetIdx int, snip *Snippet) {
	fi, folder := d.FindFolder(folderID)
	if folder == nil {
		return -1, -1, nil
	}
	for si := range folder.Snippets {
		if folder.Snippets[si].ID == snippetID {
			return fi, si, &folder.Snippets[si]
		}
	}
	return fi, -1, nil
}

// AllSnippets flattens every snippet into refs, preserving folder and snippet order.
func (d *Data) AllSnippets() []SnippetRef {
	if d == nil {
		return nil
	}
	var refs []SnippetRef
	for _, f := range d.Folders {
		for _, s := range f.Snippets {
			refs = append(refs, SnippetRef{
				FolderID:  f.ID,
				SnippetID: s.ID,
				Folder:    f.Name,
				Title:     s.Title,
				Command:   s.Command,
			})
		}
	}
	return refs
}

// RecentSnippets resolves the recent list to live snippet refs (max RecentLimit).
// Stale entries that no longer exist are skipped.
func (d *Data) RecentSnippets() []SnippetRef {
	if d == nil || len(d.Recent) == 0 {
		return nil
	}
	out := make([]SnippetRef, 0, min(len(d.Recent), RecentLimit))
	for _, e := range d.Recent {
		if len(out) >= RecentLimit {
			break
		}
		_, _, snip := d.FindSnippet(e.FolderID, e.SnippetID)
		if snip == nil {
			continue
		}
		_, folder := d.FindFolder(e.FolderID)
		name := ""
		if folder != nil {
			name = folder.Name
		}
		out = append(out, SnippetRef{
			FolderID:  e.FolderID,
			SnippetID: e.SnippetID,
			Folder:    name,
			Title:     snip.Title,
			Command:   snip.Command,
		})
	}
	return out
}

// TouchRecent records a snippet as most-recently used and trims to RecentLimit.
func (d *Data) TouchRecent(folderID, snippetID string) {
	if d == nil || folderID == "" || snippetID == "" || folderID == RecentFolderID {
		return
	}
	next := make([]RecentEntry, 0, RecentLimit)
	next = append(next, RecentEntry{FolderID: folderID, SnippetID: snippetID})
	for _, e := range d.Recent {
		if e.FolderID == folderID && e.SnippetID == snippetID {
			continue
		}
		next = append(next, e)
		if len(next) >= RecentLimit {
			break
		}
	}
	d.Recent = next
}

// MoveFolder swaps a folder with its neighbor by delta (-1 or +1).
func (d *Data) MoveFolder(folderID string, delta int) bool {
	i, _ := d.FindFolder(folderID)
	if i < 0 {
		return false
	}
	j := i + delta
	if j < 0 || j >= len(d.Folders) {
		return false
	}
	d.Folders[i], d.Folders[j] = d.Folders[j], d.Folders[i]
	return true
}

// MoveSnippet swaps a snippet with its neighbor within a folder.
func (d *Data) MoveSnippet(folderID, snippetID string, delta int) bool {
	fi, si, _ := d.FindSnippet(folderID, snippetID)
	if fi < 0 || si < 0 {
		return false
	}
	snips := d.Folders[fi].Snippets
	sj := si + delta
	if sj < 0 || sj >= len(snips) {
		return false
	}
	snips[si], snips[sj] = snips[sj], snips[si]
	d.Folders[fi].Snippets = snips
	return true
}

// DeleteFolder removes a folder and any matching recent entries.
// Returns false if the folder does not exist.
func (d *Data) DeleteFolder(folderID string) bool {
	if d == nil || folderID == "" || folderID == RecentFolderID {
		return false
	}
	fi, _ := d.FindFolder(folderID)
	if fi < 0 {
		return false
	}
	d.Folders = append(d.Folders[:fi], d.Folders[fi+1:]...)
	d.pruneRecent(folderID, "")
	return true
}

// DeleteSnippet removes a snippet from its folder and any matching recent entry.
// Returns false if the snippet does not exist.
func (d *Data) DeleteSnippet(folderID, snippetID string) bool {
	if d == nil || folderID == "" || snippetID == "" {
		return false
	}
	fi, si, _ := d.FindSnippet(folderID, snippetID)
	if fi < 0 || si < 0 {
		return false
	}
	snips := d.Folders[fi].Snippets
	d.Folders[fi].Snippets = append(snips[:si], snips[si+1:]...)
	d.pruneRecent(folderID, snippetID)
	return true
}

// pruneRecent drops recent entries for a folder (snippetID empty) or one snippet.
func (d *Data) pruneRecent(folderID, snippetID string) {
	if d == nil || len(d.Recent) == 0 {
		return
	}
	next := d.Recent[:0]
	for _, e := range d.Recent {
		if e.FolderID != folderID {
			next = append(next, e)
			continue
		}
		if snippetID != "" && e.SnippetID != snippetID {
			next = append(next, e)
		}
	}
	d.Recent = next
}

// AddFolder appends a folder with a unique id derived from name.
// Returns the new folder index within Data.Folders (not including virtual Recent).
func (d *Data) AddFolder(name string) (int, error) {
	if d == nil {
		return -1, fmt.Errorf("data is nil")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return -1, fmt.Errorf("folder name is empty")
	}
	id := uniqueID(slugify(name), func(id string) bool {
		if id == RecentFolderID {
			return true
		}
		_, f := d.FindFolder(id)
		return f != nil
	})
	d.Folders = append(d.Folders, Folder{
		ID:       id,
		Name:     name,
		Snippets: nil,
	})
	return len(d.Folders) - 1, nil
}

// AddSnippet appends a snippet to the given folder.
// Title is derived from the first line of command when empty.
// Returns the new snippet index within the folder.
func (d *Data) AddSnippet(folderID, title, command string) (int, error) {
	if d == nil {
		return -1, fmt.Errorf("data is nil")
	}
	command = strings.TrimRight(command, "\r\n")
	if strings.TrimSpace(command) == "" {
		return -1, fmt.Errorf("command is empty")
	}
	fi, folder := d.FindFolder(folderID)
	if folder == nil {
		return -1, fmt.Errorf("folder %q not found", folderID)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = titleFromCommand(command)
	}
	id := uniqueID(slugify(title), func(id string) bool {
		for _, s := range folder.Snippets {
			if s.ID == id {
				return true
			}
		}
		return false
	})
	folder.Snippets = append(folder.Snippets, Snippet{
		ID:      id,
		Title:   title,
		Command: command,
	})
	d.Folders[fi] = *folder
	return len(folder.Snippets) - 1, nil
}

func titleFromCommand(command string) string {
	first := command
	if i := strings.IndexByte(command, '\n'); i >= 0 {
		first = command[:i]
	}
	first = strings.TrimSpace(first)
	if first == "" {
		return "untitled"
	}
	runes := []rune(first)
	if len(runes) > 48 {
		return string(runes[:48]) + "…"
	}
	return first
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "item"
	}
	if len(out) > 48 {
		out = out[:48]
		out = strings.Trim(out, "-")
	}
	return out
}

func uniqueID(base string, taken func(string) bool) string {
	if base == "" {
		base = "item"
	}
	if !taken(base) {
		return base
	}
	for i := 2; ; i++ {
		id := fmt.Sprintf("%s-%d", base, i)
		if !taken(id) {
			return id
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
