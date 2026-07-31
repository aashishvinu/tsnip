// Package app owns the Bubble Tea model: state, updates, and orchestration.
// UI packages never touch storage; this package is the mediator.
package app

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aashishvinu/tsnip/internal/clipboard"
	"github.com/aashishvinu/tsnip/internal/components"
	"github.com/aashishvinu/tsnip/internal/keymap"
	"github.com/aashishvinu/tsnip/internal/models"
	"github.com/aashishvinu/tsnip/internal/search"
	"github.com/aashishvinu/tsnip/internal/storage"
	"github.com/aashishvinu/tsnip/internal/ui"
)

// Focus identifies which panel receives navigation keys.
type Focus int

const (
	FocusFolders Focus = iota
	FocusSnippets
)

type inputMode int

const (
	modeBrowse inputMode = iota
	modeNewFolder
	modeNewSnippet
)

// Model is the root Bubble Tea model for tsnip.
type Model struct {
	data   *models.Data
	store  storage.Store
	engine search.Engine
	keys   keymap.KeyMap

	width  int
	height int

	focus      Focus
	navIdx     int // sidebar: 0 = Recent, 1.. = folders
	snippetIdx int
	query      string
	searching  bool // true while filtering; typing letters/digits activates it
	results    []models.SnippetRef

	selectedCmd string
	quitting    bool
	cancelled   bool
	persistErr  string
	statusMsg   string

	mode  inputMode
	draft string

	// Mouse double-click tracking for execute-on-second-click.
	lastClickPanel ui.Panel
	lastClickIndex int
	lastClickAt    time.Time
}

// Config wires dependencies into a new Model.
type Config struct {
	Data   *models.Data
	Store  storage.Store
	Engine search.Engine
}

// New creates the application model.
func New(cfg Config) Model {
	engine := cfg.Engine
	if engine == nil {
		engine = search.New()
	}
	m := Model{
		data:   cfg.Data,
		store:  cfg.Store,
		engine: engine,
		keys:   keymap.Default(),
		focus:  FocusSnippets,
		navIdx: 0, // Recent
	}
	m.refreshVisible()
	return m
}

// SelectedCommand returns the command chosen with Enter, if any.
func (m Model) SelectedCommand() string { return m.selectedCmd }

// Cancelled reports whether the user quit without selecting.
func (m Model) Cancelled() bool { return m.cancelled || m.selectedCmd == "" }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode != modeBrowse {
		return m.handleCreateKey(msg)
	}
	return m.handleBrowseKey(msg)
}

func (m Model) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.cancelled = true
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, m.keys.Escape):
		if m.searching || m.query != "" {
			m.searching = false
			m.query = ""
			m.snippetIdx = 0
			m.statusMsg = ""
			m.refreshVisible()
			return m, nil
		}
		m.cancelled = true
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, m.keys.Copy), msg.String() == " ", msg.String() == "space", msg.Type == tea.KeySpace:
		// While searching, space is a literal character in the query — never copy.
		if m.searching {
			m.query += " "
			m.snippetIdx = 0
			m.statusMsg = ""
			m.refreshVisible()
			return m, nil
		}
		return m, m.copyAndQuit()

	case key.Matches(msg, m.keys.Search):
		m.searching = true
		m.focus = FocusSnippets
		m.statusMsg = ""
		return m, nil

	case key.Matches(msg, m.keys.NewFolder):
		m.beginCreate(modeNewFolder)
		return m, nil

	case key.Matches(msg, m.keys.NewSnippet):
		if m.data == nil || len(m.data.Folders) == 0 {
			m.beginCreate(modeNewFolder)
			return m, nil
		}
		m.beginCreate(modeNewSnippet)
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		m.deleteFocused()
		return m, nil

	case key.Matches(msg, m.keys.ClearSearch):
		m.query = ""
		m.searching = false
		m.snippetIdx = 0
		m.refreshVisible()
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		return m, m.executeSelected()

	case key.Matches(msg, m.keys.MoveUp):
		m.reorder(-1)
		return m, nil

	case key.Matches(msg, m.keys.MoveDown):
		m.reorder(1)
		return m, nil

	case key.Matches(msg, m.keys.Up):
		m.moveSelection(-1)
		return m, nil

	case key.Matches(msg, m.keys.Down):
		m.moveSelection(1)
		return m, nil

	case key.Matches(msg, m.keys.Left), key.Matches(msg, m.keys.PrevPanel):
		// Always jump to folder panel (exit search filter view for browsing).
		if m.searching {
			m.searching = false
			m.query = ""
			m.refreshVisible()
		}
		m.focus = FocusFolders
		m.statusMsg = ""
		return m, nil

	case key.Matches(msg, m.keys.Right), key.Matches(msg, m.keys.NextPanel):
		m.focus = FocusSnippets
		m.statusMsg = ""
		return m, nil
	}

	// Typing letters/digits starts search immediately (no need for / first).
	if msg.Type == tea.KeyBackspace || msg.String() == "ctrl+h" {
		if m.searching || m.query != "" {
			if m.query != "" {
				r := []rune(m.query)
				m.query = string(r[:len(r)-1])
				m.snippetIdx = 0
				m.statusMsg = ""
				m.refreshVisible()
			}
			if m.query == "" {
				m.searching = false
			}
			return m, nil
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		added := false
		for _, r := range msg.Runes {
			if !unicode.IsPrint(r) || unicode.IsControl(r) || r == ' ' {
				continue
			}
			if !m.searching {
				// Only letters/digits kick off search from browse mode.
				if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
					continue
				}
				m.searching = true
			}
			m.query += string(r)
			added = true
		}
		if added {
			m.snippetIdx = 0
			m.focus = FocusSnippets
			m.statusMsg = ""
			m.refreshVisible()
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) executeSelected() tea.Cmd {
	ref, ok := m.currentRef()
	if !ok {
		return nil
	}
	m.touchRecent(ref)
	m.selectedCmd = ref.Command
	m.cancelled = false
	m.quitting = true
	return tea.Quit
}

func (m *Model) copyAndQuit() tea.Cmd {
	if !m.copyCurrent() {
		return nil
	}
	m.cancelled = true
	m.quitting = true
	return tea.Quit
}

const doubleClickWindow = 400 * time.Millisecond

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.mode != modeBrowse || m.width <= 0 || m.height <= 0 {
		return m, nil
	}

	layout := ui.ComputeLayout(m.width, m.height)
	hit := layout.HitTest(msg.X, msg.Y)

	switch msg.Button {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		delta := 1
		if msg.Button == tea.MouseButtonWheelUp {
			delta = -1
		}
		switch hit.Panel {
		case ui.PanelFolders:
			if m.searching {
				m.searching = false
				m.query = ""
				m.refreshVisible()
			}
			m.focus = FocusFolders
			m.moveSelection(delta)
		case ui.PanelSnippets, ui.PanelSearch, ui.PanelNone:
			m.focus = FocusSnippets
			m.moveSelection(delta)
		}
		return m, nil
	}

	if msg.Action != tea.MouseActionPress {
		return m, nil
	}

	folders, snippets := m.panelLists(layout)

	switch msg.Button {
	case tea.MouseButtonLeft:
		switch hit.Panel {
		case ui.PanelSearch:
			m.searching = true
			m.focus = FocusSnippets
			m.statusMsg = ""
			m.lastClickPanel = ui.PanelSearch
			m.lastClickIndex = -1
			m.lastClickAt = time.Now()
			return m, nil

		case ui.PanelFolders:
			idx := folders.IndexAt(hit.Row)
			if idx < 0 {
				return m, nil
			}
			if m.searching {
				m.searching = false
				m.query = ""
			}
			m.focus = FocusFolders
			m.navIdx = idx
			m.snippetIdx = 0
			m.statusMsg = ""
			m.refreshVisible()
			m.lastClickPanel = ui.PanelFolders
			m.lastClickIndex = idx
			m.lastClickAt = time.Now()
			return m, nil

		case ui.PanelSnippets:
			idx := snippets.IndexAt(hit.Row)
			if idx < 0 {
				m.focus = FocusSnippets
				return m, nil
			}
			m.focus = FocusSnippets
			m.snippetIdx = idx
			m.statusMsg = ""
			m.lastClickPanel = ui.PanelSnippets
			m.lastClickIndex = idx
			m.lastClickAt = time.Now()
			return m, m.executeSelected()
		}

	case tea.MouseButtonRight:
		if hit.Panel == ui.PanelSnippets {
			idx := snippets.IndexAt(hit.Row)
			if idx >= 0 {
				m.focus = FocusSnippets
				m.snippetIdx = idx
			}
			return m, m.copyAndQuit()
		}
		if hit.Panel == ui.PanelFolders {
			idx := folders.IndexAt(hit.Row)
			if idx >= 0 {
				if m.searching {
					m.searching = false
					m.query = ""
				}
				m.focus = FocusFolders
				m.navIdx = idx
				m.snippetIdx = 0
				m.refreshVisible()
			}
		}
	}

	return m, nil
}

func (m Model) panelLists(layout ui.Layout) (components.FolderList, components.SnippetList) {
	folderItems := []components.FolderItem{{Name: "Recent", Special: true}}
	if m.data != nil {
		for _, f := range m.data.Folders {
			folderItems = append(folderItems, components.FolderItem{Name: f.Name})
		}
	}

	items := make([]components.SnippetItem, len(m.results))
	for i, r := range m.results {
		items[i] = components.SnippetItem{Code: r.Command}
	}

	return components.FolderList{
			Items:  folderItems,
			Cursor: m.navIdx,
			Width:  layout.FolderW,
			Height: layout.BodyH,
		}, components.SnippetList{
			Items:  items,
			Cursor: m.snippetIdx,
			Width:  layout.SnippetW,
			Height: layout.BodyH,
		}
}

func (m *Model) copyCurrent() bool {
	ref, ok := m.currentRef()
	if !ok {
		m.statusMsg = "nothing to copy"
		return false
	}
	if err := clipboard.Write(ref.Command); err != nil {
		m.statusMsg = "copy failed"
		return false
	}
	m.touchRecent(ref)
	m.statusMsg = "copied"
	return true
}

func (m *Model) touchRecent(ref models.SnippetRef) {
	if m.data == nil {
		return
	}
	m.data.TouchRecent(ref.FolderID, ref.SnippetID)
	m.persist()
}

func (m Model) handleCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.cancelled = true
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, m.keys.Escape):
		m.cancelCreate()
		return m, nil

	case key.Matches(msg, m.keys.Save):
		if m.mode == modeNewSnippet {
			m.commitCreate()
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if m.mode == modeNewFolder {
			m.commitCreate()
			return m, nil
		}
		m.draft += "\n"
		return m, nil

	case key.Matches(msg, m.keys.ClearSearch):
		m.draft = ""
		return m, nil
	}

	if msg.Type == tea.KeyBackspace || msg.String() == "ctrl+h" {
		if m.draft != "" {
			r := []rune(m.draft)
			m.draft = string(r[:len(r)-1])
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		for _, r := range msg.Runes {
			if unicode.IsPrint(r) && !unicode.IsControl(r) {
				m.draft += string(r)
			}
		}
		return m, nil
	}

	if msg.String() == " " {
		m.draft += " "
		return m, nil
	}

	return m, nil
}

func (m *Model) beginCreate(mode inputMode) {
	m.mode = mode
	m.draft = ""
	m.query = ""
	m.searching = false
	m.persistErr = ""
	m.statusMsg = ""
	if mode == modeNewSnippet {
		m.focus = FocusSnippets
	} else {
		m.focus = FocusFolders
	}
}

func (m *Model) cancelCreate() {
	m.mode = modeBrowse
	m.draft = ""
	m.refreshVisible()
}

func (m *Model) commitCreate() {
	switch m.mode {
	case modeNewFolder:
		idx, err := m.data.AddFolder(m.draft)
		if err != nil {
			m.persistErr = err.Error()
			return
		}
		m.navIdx = idx + 1
		m.snippetIdx = 0
		m.focus = FocusFolders
		m.persist()
		m.mode = modeBrowse
		m.draft = ""
		m.refreshVisible()

	case modeNewSnippet:
		folderID := m.targetFolderID()
		if folderID == "" {
			m.persistErr = "create a folder first"
			return
		}
		si, err := m.data.AddSnippet(folderID, "", m.draft)
		if err != nil {
			m.persistErr = err.Error()
			return
		}
		if fi, _ := m.data.FindFolder(folderID); fi >= 0 {
			m.navIdx = fi + 1
		}
		m.snippetIdx = si
		m.focus = FocusSnippets
		m.persist()
		m.mode = modeBrowse
		m.draft = ""
		m.refreshVisible()
	}
}

func (m *Model) targetFolderID() string {
	if m.data == nil || len(m.data.Folders) == 0 {
		return ""
	}
	if !m.onRecent() {
		fi := m.navIdx - 1
		if fi >= 0 && fi < len(m.data.Folders) {
			return m.data.Folders[fi].ID
		}
	}
	return m.data.Folders[0].ID
}

func (m Model) onRecent() bool { return m.navIdx == 0 }

func (m Model) sidebarLen() int {
	n := 1
	if m.data != nil {
		n += len(m.data.Folders)
	}
	return n
}

func (m *Model) moveSelection(delta int) {
	// While searching, arrows always move through results.
	if m.searching || m.focus == FocusSnippets {
		n := len(m.results)
		if n == 0 {
			return
		}
		m.focus = FocusSnippets
		m.snippetIdx = clamp(m.snippetIdx+delta, 0, n-1)
		return
	}

	n := m.sidebarLen()
	if n == 0 {
		return
	}
	m.navIdx = clamp(m.navIdx+delta, 0, n-1)
	m.snippetIdx = 0
	m.refreshVisible()
}

func (m *Model) reorder(delta int) {
	if m.data == nil || m.store == nil || m.searching || m.query != "" {
		return
	}

	if m.focus == FocusFolders {
		if m.onRecent() {
			return
		}
		fi := m.navIdx - 1
		if fi < 0 || fi >= len(m.data.Folders) {
			return
		}
		id := m.data.Folders[fi].ID
		if !m.data.MoveFolder(id, delta) {
			return
		}
		m.navIdx = clamp(m.navIdx+delta, 1, len(m.data.Folders))
		m.persist()
		m.refreshVisible()
		return
	}

	if m.onRecent() {
		return
	}
	if m.snippetIdx < 0 || m.snippetIdx >= len(m.results) {
		return
	}
	ref := m.results[m.snippetIdx]
	if !m.data.MoveSnippet(ref.FolderID, ref.SnippetID, delta) {
		return
	}
	m.snippetIdx = clamp(m.snippetIdx+delta, 0, len(m.results)-1)
	m.persist()
	m.refreshVisible()
}

func (m *Model) deleteFocused() {
	if m.data == nil || m.store == nil {
		return
	}

	// Snippets panel (or search): delete the highlighted snippet.
	if m.searching || m.focus == FocusSnippets {
		ref, ok := m.currentRef()
		if !ok {
			m.statusMsg = "nothing to delete"
			return
		}
		if !m.data.DeleteSnippet(ref.FolderID, ref.SnippetID) {
			m.statusMsg = "delete failed"
			return
		}
		m.persist()
		m.refreshVisible()
		if m.snippetIdx >= len(m.results) && m.snippetIdx > 0 {
			m.snippetIdx = len(m.results) - 1
		}
		m.statusMsg = "deleted"
		return
	}

	// Folders panel: delete the highlighted folder (not Recent).
	if m.onRecent() {
		m.statusMsg = "can't delete Recent"
		return
	}
	fi := m.navIdx - 1
	if fi < 0 || fi >= len(m.data.Folders) {
		m.statusMsg = "nothing to delete"
		return
	}
	name := m.data.Folders[fi].Name
	if !m.data.DeleteFolder(m.data.Folders[fi].ID) {
		m.statusMsg = "delete failed"
		return
	}
	if m.navIdx > len(m.data.Folders) {
		m.navIdx = len(m.data.Folders) // clamp onto last folder or Recent
	}
	m.snippetIdx = 0
	m.persist()
	m.refreshVisible()
	m.statusMsg = "deleted " + name
}

func (m *Model) persist() {
	if m.store == nil {
		return
	}
	if err := m.store.Save(m.data); err != nil {
		m.persistErr = err.Error()
		return
	}
	m.persistErr = ""
}

func (m *Model) refreshVisible() {
	m.results = m.results[:0]
	if m.data == nil {
		return
	}

	m.navIdx = clamp(m.navIdx, 0, max(0, m.sidebarLen()-1))
	corpus := m.data.AllSnippets()
	q := search.Query{Text: m.query}

	if m.searching && m.query != "" {
		for _, h := range m.engine.Search(q, corpus) {
			m.results = append(m.results, h.Ref)
		}
	} else if m.onRecent() {
		m.results = append(m.results, m.data.RecentSnippets()...)
	} else {
		fi := m.navIdx - 1
		if fi >= 0 && fi < len(m.data.Folders) {
			q.FolderID = m.data.Folders[fi].ID
			q.Text = ""
			for _, h := range m.engine.Search(q, corpus) {
				m.results = append(m.results, h.Ref)
			}
		}
	}

	if len(m.results) == 0 {
		m.snippetIdx = 0
		return
	}
	m.snippetIdx = clamp(m.snippetIdx, 0, len(m.results)-1)
}

func (m Model) currentRef() (models.SnippetRef, bool) {
	if m.snippetIdx < 0 || m.snippetIdx >= len(m.results) {
		return models.SnippetRef{}, false
	}
	return m.results[m.snippetIdx], true
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	folderItems := []components.FolderItem{{Name: "Recent", Special: true}}
	if m.data != nil {
		for _, f := range m.data.Folders {
			folderItems = append(folderItems, components.FolderItem{Name: f.Name})
		}
	}

	filtering := m.searching && m.query != ""
	composing := m.mode != modeBrowse

	items := make([]components.SnippetItem, len(m.results))
	for i, r := range m.results {
		items[i] = components.SnippetItem{Code: r.Command}
	}

	emptyMsg := "no snippets"
	switch {
	case filtering:
		emptyMsg = "no matches"
	case m.onRecent():
		emptyMsg = "no recent snippets yet"
	}

	searchPrompt := "/"
	searchPlaceholder := "search snippets"
	searchQuery := m.query
	searchRight := m.contextLabel()
	if m.searching {
		searchPlaceholder = "filter…"
	}

	switch m.mode {
	case modeNewFolder:
		searchPrompt = "folder"
		searchPlaceholder = "name"
		searchQuery = m.draft
		searchRight = "enter save · esc cancel"
	case modeNewSnippet:
		searchPrompt = "snippet"
		searchPlaceholder = "command · enter newline"
		searchQuery = m.draft
		searchRight = "^s save · esc cancel"
	}

	if m.persistErr != "" {
		searchRight = m.persistErr
	} else if m.statusMsg != "" {
		searchRight = m.statusMsg
	}

	return ui.Model{
		Width:  m.width,
		Height: m.height,
		Folders: components.FolderList{
			Items:    folderItems,
			Cursor:   m.navIdx,
			Focused:  m.focus == FocusFolders && !m.searching && !composing,
			Disabled: m.searching || m.mode == modeNewSnippet,
		},
		Snippets: components.SnippetList{
			Items:   items,
			Cursor:  m.snippetIdx,
			Focused: (m.focus == FocusSnippets || m.searching) && !composing,
			Empty:   emptyMsg,
		},
		Search: components.SearchBar{
			Prompt:      searchPrompt,
			Placeholder: searchPlaceholder,
			Query:       searchQuery,
			Cursor:      m.searching || composing,
			Active:      m.searching || composing,
			Right:       searchRight,
		},
		Searching:    m.searching,
		Composing:    m.mode == modeNewSnippet,
		ComposeCode:  strings.TrimRight(m.draft, "\r"),
		ComposeLabel: "type a command — enter for newline",
	}.View()
}

func (m Model) contextLabel() string {
	if m.data == nil {
		return ""
	}
	if m.searching {
		if m.query == "" {
			return "type to filter"
		}
		n := len(m.results)
		if n == 1 {
			return "1 match"
		}
		return fmt.Sprintf("%d matches", n)
	}
	if m.onRecent() {
		return fmt.Sprintf("%d recent", len(m.results))
	}
	fi := m.navIdx - 1
	if fi >= 0 && fi < len(m.data.Folders) {
		return fmt.Sprintf("%d snippets", len(m.results))
	}
	return ""
}

func (m Model) targetFolderName() string {
	id := m.targetFolderID()
	if id == "" {
		return ""
	}
	_, f := m.data.FindFolder(id)
	if f == nil {
		return ""
	}
	return f.Name
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
