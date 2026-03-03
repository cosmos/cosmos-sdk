package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

// viewModel is implemented by every per-view type.
type viewModel interface {
	Update(msg tea.Msg) (viewModel, tea.Cmd)
	View() string
	Title() string
	KeyMap() help.KeyMap
}

// pushViewMsg is returned by views to push a new child view onto the stack.
type pushViewMsg struct{ view viewModel }

func pushView(v viewModel) tea.Cmd {
	return func() tea.Msg { return pushViewMsg{v} }
}

// errMsg is returned by views to display an error in the root.
type errMsg struct{ err error }

// globalKeyMap defines keybindings available on every view.
type globalKeyMap struct {
	Back    key.Binding
	Quit    key.Binding
	Refresh key.Binding
	Help    key.Binding
}

func newGlobalKeyMap() globalKeyMap {
	return globalKeyMap{
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// refreshable is optionally implemented by views that can reload data from disk.
type refreshable interface {
	refresh()
}

// combinedKeyMap merges view-specific keys with global keys for the help footer.
type combinedKeyMap struct {
	view        help.KeyMap
	global      globalKeyMap
	showBack    bool
	showRefresh bool
}

func (c combinedKeyMap) ShortHelp() []key.Binding {
	bindings := c.view.ShortHelp()
	if c.showBack {
		bindings = append(bindings, c.global.Back)
	}
	if c.showRefresh {
		bindings = append(bindings, c.global.Refresh)
	}
	bindings = append(bindings, c.global.Quit)
	bindings = append(bindings, c.global.Help)
	return bindings
}

func (c combinedKeyMap) FullHelp() [][]key.Binding {
	groups := c.view.FullHelp()
	var global []key.Binding
	if c.showBack {
		global = append(global, c.global.Back)
	}
	if c.showRefresh {
		global = append(global, c.global.Refresh)
	}
	global = append(global, c.global.Quit)
	groups = append(groups, global)
	return groups
}

// model is the root Bubble Tea model with a stack of views.
type model struct {
	stack     []viewModel
	width     int
	height    int
	help      help.Model
	keys      globalKeyMap
	err       string
	helpModal helpModal
}

var tableStyles table.Styles

func init() {
	tableStyles = table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	tableStyles.Selected = tableStyles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
}

func initialModel(dir string) model {
	h := help.New()
	m := model{
		keys: newGlobalKeyMap(),
		help: h,
	}
	v := newTreesView(dir, 18)
	m.stack = []viewModel{v}
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pushViewMsg:
		m.stack = append(m.stack, msg.view)
		m.err = ""
		// Send a synthetic resize so the new view knows content dimensions.
		if m.width > 0 {
			var cmd tea.Cmd
			top := m.stack[len(m.stack)-1]
			top, cmd = top.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.stack[len(m.stack)-1] = top
			return m, cmd
		}
		return m, nil

	case errMsg:
		m.err = msg.err.Error()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.helpModal.resize(msg.Width, msg.Height)
		// Delegate to current view.
		top := m.stack[len(m.stack)-1]
		var cmd tea.Cmd
		top, cmd = top.Update(msg)
		m.stack[len(m.stack)-1] = top
		return m, cmd

	case tea.KeyMsg:
		// When modal is open, intercept all keys.
		if m.helpModal.visible {
			switch {
			case key.Matches(msg, m.keys.Help), key.Matches(msg, m.keys.Back):
				m.helpModal.visible = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.helpModal.viewport, cmd = m.helpModal.viewport.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Back):
			if len(m.stack) <= 1 {
				return m, nil
			}
			m.stack = m.stack[:len(m.stack)-1]
			m.err = ""
			// Re-send resize to restored view.
			if m.width > 0 {
				top := m.stack[len(m.stack)-1]
				var cmd tea.Cmd
				top, cmd = top.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				m.stack[len(m.stack)-1] = top
				return m, cmd
			}
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			top := m.stack[len(m.stack)-1]
			if r, ok := top.(refreshable); ok {
				r.refresh()
				m.err = ""
			}
			return m, nil
		case key.Matches(msg, m.keys.Help):
			top := m.stack[len(m.stack)-1]
			doc := appHelpDoc
			if d, ok := top.(helpDocer); ok {
				doc = d.HelpDoc()
			}
			m.helpModal.open(doc, m.width, m.height)
			return m, nil
		}
	}

	// Delegate everything else to the active view.
	top := m.stack[len(m.stack)-1]
	var cmd tea.Cmd
	top, cmd = top.Update(msg)
	m.stack[len(m.stack)-1] = top
	return m, cmd
}

func (m model) View() string {
	if m.helpModal.visible {
		return m.helpModal.render(m.width, m.height)
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	top := m.stack[len(m.stack)-1]
	title := boxStyle.Render(top.Title())
	_, topIsRefreshable := top.(refreshable)
	footer := boxStyle.Render(m.help.View(combinedKeyMap{view: top.KeyMap(), global: m.keys, showBack: len(m.stack) > 1, showRefresh: topIsRefreshable}))

	if m.err != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1)
		return title + "\n" + errStyle.Render("Error: "+m.err) + "\n" + footer
	}

	return title + "\n" + top.View() + "\n" + footer
}

var _ tea.Model = model{}

// safeTableKeyMap returns a table.KeyMap that avoids single-letter conflicts.
func safeTableKeyMap() table.KeyMap {
	return table.KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "go to end"),
		),
	}
}

func newTable(columns []table.Column, rows []table.Row, height int) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
		table.WithKeyMap(safeTableKeyMap()),
	)
	t.SetStyles(tableStyles)
	return t
}

func contentHeight(totalHeight int) int {
	if totalHeight > 0 {
		return totalHeight - 10
	}
	return 18
}

func statSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func fmtCountAndSize(size, structSize int64) string {
	if size == 0 {
		return "-"
	}
	if structSize > 0 {
		return fmt.Sprintf("%d (%s)", size/structSize, humanSize(size))
	}
	return humanSize(size)
}

func fmtOrphanPct(orphanSize, leafSize, branchSize int64) string {
	orphanCount := orphanSize / internal.SizeOrphanEntry
	nodeCount := leafSize/internal.SizeLeaf + branchSize/internal.SizeBranch
	if nodeCount == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", float64(orphanCount)/float64(nodeCount)*100)
}

type sizeEntry struct {
	label string
	size  int64
}

func renderSizeBreakdown(sizes []sizeEntry) string {
	var total int64
	for _, e := range sizes {
		total += e.size
	}
	if total == 0 {
		return ""
	}

	sort.Slice(sizes, func(i, j int) bool {
		return sizes[i].size > sizes[j].size
	})

	const barWidth = 30
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	var b strings.Builder
	b.WriteString(dimStyle.Render(fmt.Sprintf("  Size Breakdown (%s total)", humanSize(total))))
	b.WriteByte('\n')

	for _, e := range sizes {
		if e.size == 0 {
			continue
		}
		pct := float64(e.size) * 100.0 / float64(total)
		filled := float64(barWidth) * float64(e.size) / float64(total)
		fullBlocks := int(filled)
		frac := filled - float64(fullBlocks)

		bar := strings.Repeat("█", fullBlocks)
		if frac > 0.0625 {
			fractional := []rune{'▏', '▎', '▍', '▌', '▋', '▊', '▉'}
			idx := int(frac * 8)
			if idx > 6 {
				idx = 6
			}
			bar += string(fractional[idx])
		}
		if bar == "" {
			bar = "▏"
		}

		b.WriteString(fmt.Sprintf("  %-12s %s %5.1f%%  %s\n",
			e.label,
			barStyle.Render(fmt.Sprintf("%-*s", barWidth, bar)),
			pct,
			humanSize(e.size),
		))
	}
	return b.String()
}

func formatStat(s *runningStats, f func(*runningStats) float64) string {
	if s.n == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f", f(s))
}

type orphanCounts struct {
	leaves   int
	branches int
}

func renderInfoPanel(columns []table.Column, row table.Row, width int) string {
	if row == nil || len(columns) == 0 {
		return ""
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	if width <= 0 {
		width = 120
	}

	sep := "  "
	var line string
	for i, col := range columns {
		if i >= len(row) {
			break
		}
		pair := dimStyle.Render(col.Title+":") + " " + row[i]
		candidate := pair
		if line != "" {
			candidate = line + sep + pair
		}
		if lipgloss.Width(candidate) > width {
			break
		}
		line = candidate
	}

	rule := dimStyle.Render(strings.Repeat("─", width))
	return rule + "\n" + line + "\n"
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// emptyKeyMap is used by views with no action keys.
type emptyKeyMap struct{}

func (emptyKeyMap) ShortHelp() []key.Binding  { return nil }
func (emptyKeyMap) FullHelp() [][]key.Binding { return nil }

// isTotalRow returns true if the first column starts with the TOTAL marker.
func isTotalRow(row table.Row) bool {
	return row != nil && strings.HasPrefix(row[0], "━━")
}
