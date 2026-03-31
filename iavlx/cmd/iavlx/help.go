package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/cosmos/cosmos-sdk/iavlx/docs"
)

// glamourStyle is detected once at startup, before BubbleTea takes over the
// terminal. glamour.WithAutoStyle() re-queries the terminal on every call via
// termenv OSC 11, which hangs up to 5 s when BubbleTea is in raw mode.
var glamourStyle = func() string {
	if lipgloss.HasDarkBackground() {
		return "dark"
	}
	return "light"
}()

// helpDocer is optionally implemented by views to identify their help topic.
// The returned key is the .md filename in the embedded docs (e.g. "checkpoint.md").
type helpDocer interface {
	HelpDocKey() string
}

// docEntry is one entry in the help sidebar TOC.
// key is the .md filename; content is loaded on demand via loadDocMarkdown.
type docEntry struct {
	key   string
	label string
}

const sidebarW = 16

// allDocs is the fixed list of help topics in sidebar order.
var allDocs = []docEntry{
	{key: "commit-lifecycle.md", label: "Commit"},
	{key: "multitree.md", label: "Multi-Tree"},
	{key: "changeset.md", label: "Changesets"},
	{key: "checkpoint.md", label: "Checkpoints"},
	{key: "leaves.md", label: "Leaves"},
	{key: "branches.md", label: "Branches"},
	{key: "node-id.md", label: "Node IDs"},
	{key: "orphans.md", label: "Orphans"},
	{key: "wal.md", label: "WAL"},
	{key: "commit-info.md", label: "Commit Info"},
}

// modalSidebarKeyMap is the help footer shown when the sidebar has focus.
type modalSidebarKeyMap struct{}

func (modalSidebarKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "content pane")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "navigate")),
		key.NewBinding(key.WithKeys("esc", "?"), key.WithHelp("esc/?", "close")),
		key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
func (modalSidebarKeyMap) FullHelp() [][]key.Binding { return nil }

// modalContentKeyMap is the help footer shown when the content pane has focus.
type modalContentKeyMap struct{}

func (modalContentKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "topics")),
		key.NewBinding(key.WithKeys("up", "down", "pgup", "pgdown"), key.WithHelp("↑/↓/pgup/pgdn", "scroll")),
		key.NewBinding(key.WithKeys("esc", "?"), key.WithHelp("esc/?", "close")),
		key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
func (modalContentKeyMap) FullHelp() [][]key.Binding { return nil }

// helpModal holds state for the floating help documentation modal.
type helpModal struct {
	visible        bool
	viewport       viewport.Model
	sidebarCursor  int
	sidebarFocused bool
	renderCache    map[int]string
	cacheVpW       int
	width          int
	height         int
}

// helpRenderedMsg is sent when background glamour rendering completes.
type helpRenderedMsg struct {
	idx      int
	vpW      int
	rendered string
}

func (h *helpModal) vpDims(totalW, totalH int) (vpW, vpH int) {
	// Inner content width = totalW - 6 (border 1+1, padding 2+2).
	// Split as: sidebar(sidebarW) + divider(1) + viewport(rest).
	vpW = totalW - 6 - sidebarW - 1
	if vpW < 20 {
		vpW = 20
	}
	// border(2) + title(1) + footer(1) = 4 lines overhead
	vpH = totalH - 4
	if vpH < 5 {
		vpH = 5
	}
	return
}

// open makes the modal visible immediately and returns a Cmd to render the
// active doc in the background.
func (h *helpModal) open(docKey string, totalW, totalH int) tea.Cmd {
	h.visible = true
	h.width = totalW
	h.height = totalH
	h.sidebarFocused = false
	if h.renderCache == nil {
		h.renderCache = make(map[int]string)
	}

	// Find the sidebar entry matching the key.
	h.sidebarCursor = 0
	for i, e := range allDocs {
		if e.key == docKey {
			h.sidebarCursor = i
			break
		}
	}

	vpW, vpH := h.vpDims(totalW, totalH)
	h.cacheVpW = vpW
	h.viewport = viewport.New(vpW, vpH)
	return h.loadCurrent()
}

// loadCurrent populates the viewport for the current sidebar cursor, returning
// a background render Cmd if the doc is not yet cached.
func (h *helpModal) loadCurrent() tea.Cmd {
	idx := h.sidebarCursor
	if idx < 0 || idx >= len(allDocs) {
		return nil
	}
	if rendered, ok := h.renderCache[idx]; ok {
		h.viewport.SetContent(rendered)
		h.viewport.GotoTop()
		return nil
	}
	h.viewport.SetContent("Rendering…")
	h.viewport.GotoTop()
	vpW := h.cacheVpW
	doc := loadDocMarkdown(allDocs[idx].key)
	return func() tea.Msg {
		r, err := glamour.NewTermRenderer(glamour.WithStylePath(glamourStyle), glamour.WithWordWrap(vpW))
		var rendered string
		if err == nil {
			rendered, err = r.Render(doc)
		}
		if err != nil {
			rendered = doc
		}
		return helpRenderedMsg{idx: idx, vpW: vpW, rendered: rendered}
	}
}

// setRendered is called when a background render completes.
func (h *helpModal) setRendered(msg helpRenderedMsg) {
	if msg.vpW != h.cacheVpW {
		return // stale render from before a resize; discard
	}
	h.renderCache[msg.idx] = msg.rendered
	if msg.idx == h.sidebarCursor {
		h.viewport.SetContent(msg.rendered)
		h.viewport.GotoTop()
	}
}

func (h *helpModal) resize(totalW, totalH int) tea.Cmd {
	if !h.visible {
		return nil
	}
	vpW, vpH := h.vpDims(totalW, totalH)
	if vpW != h.cacheVpW {
		h.renderCache = make(map[int]string) // invalidate cache on width change
		h.cacheVpW = vpW
	}
	h.width = totalW
	h.height = totalH
	h.viewport.Width = vpW
	h.viewport.Height = vpH
	return h.loadCurrent()
}

// handleKey routes key events inside the modal. Returns a Cmd and whether
// the modal should close.
func (h *helpModal) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "tab":
		h.sidebarFocused = !h.sidebarFocused
		return nil, false
	}

	if h.sidebarFocused {
		switch msg.String() {
		case "up", "k":
			if h.sidebarCursor > 0 {
				h.sidebarCursor--
				return h.loadCurrent(), false
			}
		case "down", "j":
			if h.sidebarCursor < len(allDocs)-1 {
				h.sidebarCursor++
				return h.loadCurrent(), false
			}
		}
		return nil, false
	}

	// Content pane focused — pass all other keys to the viewport.
	var cmd tea.Cmd
	h.viewport, cmd = h.viewport.Update(msg)
	return cmd, false
}

func (h *helpModal) render(totalW, _ int) string {
	vpH := h.viewport.Height

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("62"))

	// Sidebar column.
	var sidebarLines []string
	for i, e := range allDocs {
		label := e.label
		if len([]rune(label)) > sidebarW-2 {
			label = string([]rune(label)[:sidebarW-2])
		}
		var line string
		switch {
		case i == h.sidebarCursor && h.sidebarFocused:
			line = activeStyle.Width(sidebarW).Render("▶ " + label)
		case i == h.sidebarCursor:
			line = selectedStyle.Width(sidebarW).Render("  " + label)
		default:
			line = dimStyle.Width(sidebarW).Render("  " + label)
		}
		sidebarLines = append(sidebarLines, line)
	}
	// Pad sidebar to vpH lines.
	blank := strings.Repeat(" ", sidebarW)
	for len(sidebarLines) < vpH {
		sidebarLines = append(sidebarLines, blank)
	}
	sidebarBlock := strings.Join(sidebarLines[:vpH], "\n")

	// Divider column.
	divLines := make([]string, vpH)
	for i := range divLines {
		divLines[i] = dimStyle.Render("│")
	}
	dividerBlock := strings.Join(divLines, "\n")

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarBlock, dividerBlock, h.viewport.View())

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	title := titleStyle.Render(" Documentation ")

	hlp := help.New()
	hlp.Width = totalW - 6 // inner content width (border 2 + padding 4)
	var footer string
	if h.sidebarFocused {
		footer = hlp.View(modalSidebarKeyMap{})
	} else {
		footer = hlp.View(modalContentKeyMap{})
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 2).
		Width(totalW - 2)

	return boxStyle.Render(title + "\n" + body + "\n" + footer)
}

func loadDocMarkdown(docFileName string) string {
	data, err := docs.Docs.ReadFile(docFileName)
	if err != nil {
		return fmt.Sprintf("# %s\n\nDocumentation not found: %s", docFileName, err)
	}
	return string(data)
}
