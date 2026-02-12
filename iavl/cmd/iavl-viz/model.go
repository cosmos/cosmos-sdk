package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view int

const (
	viewTrees view = iota
	viewChangesets
)

type model struct {
	trees         []treeEntry
	view          view
	selectedTree  int
	treeTable     table.Model
	csTable       table.Model
	width, height int
	dir           string
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

func initialModel(dir string, trees []treeEntry) model {
	m := model{dir: dir, trees: trees}
	m.treeTable = m.buildTreeTable()
	return m
}

func (m model) buildTreeTable() table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Changesets", Width: 12},
		{Title: "Size", Width: 12},
	}
	rows := make([]table.Row, len(m.trees))
	for i, t := range m.trees {
		rows[i] = table.Row{t.name, strconv.Itoa(len(t.changesets)), humanSize(t.totalSize)}
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	t.SetStyles(tableStyles)
	return t
}

func (m model) buildCSTable(tree treeEntry) table.Model {
	columns := []table.Column{
		{Title: "Start", Width: 10},
		{Title: "End", Width: 10},
		{Title: "Compacted", Width: 10},
		{Title: "Size", Width: 12},
	}
	rows := make([]table.Row, len(tree.changesets))
	for i, cs := range tree.changesets {
		endStr := "-"
		if cs.end > 0 {
			endStr = strconv.FormatUint(uint64(cs.end), 10)
		}
		compStr := "-"
		if cs.compacted > 0 {
			compStr = strconv.FormatUint(uint64(cs.compacted), 10)
		}
		rows[i] = table.Row{
			strconv.FormatUint(uint64(cs.start), 10),
			endStr,
			compStr,
			humanSize(cs.size),
		}
	}
	h := 20
	if m.height > 0 {
		h = m.height - 4
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(h),
	)
	t.SetStyles(tableStyles)
	return t
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.view == viewChangesets {
				m.view = viewTrees
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			if m.view == viewTrees && len(m.trees) > 0 {
				m.selectedTree = m.treeTable.Cursor()
				m.csTable = m.buildCSTable(m.trees[m.selectedTree])
				m.view = viewChangesets
				return m, nil
			}
		case "esc":
			if m.view == viewChangesets {
				m.view = viewTrees
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := msg.Height - 4
		m.treeTable.SetHeight(h)
		m.csTable.SetHeight(h)
	}

	var cmd tea.Cmd
	switch m.view {
	case viewTrees:
		m.treeTable, cmd = m.treeTable.Update(msg)
	case viewChangesets:
		m.csTable, cmd = m.csTable.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	var title string
	var tbl string
	titleStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)

	switch m.view {
	case viewTrees:
		title = titleStyle.Render("IAVL Trees: ", m.dir)
		tbl = m.treeTable.View()
	case viewChangesets:
		name := m.trees[m.selectedTree].name
		title = titleStyle.Render(fmt.Sprintf("Changesets: %s", name))
		tbl = m.csTable.View()
	}
	return title + "\n" + tbl + "\n"
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

var _ tea.Model = model{}
