package main

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type view int

const (
	viewTrees view = iota
	viewChangesets
	viewCheckpoints
	viewLeaves
	viewBranches
)

type model struct {
	view          view
	width, height int
	table         table.Model
	err           string

	dir                string
	selectedTree       string
	selectedChangeset  string
	selectedCheckpoint uint32

	checkpoints []internal.CheckpointInfo
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
	m := model{dir: dir, view: viewTrees}
	m.buildTreesTable()
	return m
}

func (m *model) tableHeight() int {
	if m.height > 0 {
		return m.height - 4
	}
	return 20
}

func newTable(columns []table.Column, rows []table.Row, height int) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	t.SetStyles(tableStyles)
	return t
}

func (m *model) buildTreesTable() {
	names, err := scanTrees(m.dir)
	if err != nil {
		m.err = err.Error()
		return
	}
	rows := make([]table.Row, len(names))
	for i, name := range names {
		treeDir := filepath.Join(m.dir, "stores", name+".iavl")
		entries, _ := os.ReadDir(treeDir)
		csCount := 0
		var totalSize int64
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			_, _, _, valid := internal.ParseChangesetDirName(e.Name())
			if !valid {
				continue
			}
			csCount++
			csPath := filepath.Join(treeDir, e.Name())
			_ = filepath.WalkDir(csPath, func(_ string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				if info, err := d.Info(); err == nil {
					totalSize += info.Size()
				}
				return nil
			})
		}
		rows[i] = table.Row{name, strconv.Itoa(csCount), humanSize(totalSize)}
	}
	m.table = newTable([]table.Column{
		{Title: "Name", Width: 30},
		{Title: "Changesets", Width: 12},
		{Title: "Size", Width: 12},
	}, rows, m.tableHeight())
}

func (m *model) buildChangesetsTable() {
	treeDir := filepath.Join(m.dir, "stores", m.selectedTree+".iavl")
	entries, err := os.ReadDir(treeDir)
	if err != nil {
		m.err = err.Error()
		return
	}
	var rows []table.Row
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		start, end, compacted, valid := internal.ParseChangesetDirName(e.Name())
		if !valid {
			continue
		}
		var size int64
		csPath := filepath.Join(treeDir, e.Name())
		_ = filepath.WalkDir(csPath, func(_ string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if info, err := d.Info(); err == nil {
				size += info.Size()
			}
			return nil
		})
		endStr := "-"
		if end > 0 {
			endStr = strconv.FormatUint(uint64(end), 10)
		}
		compStr := "-"
		if compacted > 0 {
			compStr = strconv.FormatUint(uint64(compacted), 10)
		}
		rows = append(rows, table.Row{e.Name(), strconv.FormatUint(uint64(start), 10), endStr, compStr, humanSize(size)})
	}
	m.table = newTable([]table.Column{
		{Title: "Dir", Width: 20},
		{Title: "Start", Width: 10},
		{Title: "End", Width: 10},
		{Title: "Compacted", Width: 10},
		{Title: "Size", Width: 12},
	}, rows, m.tableHeight())
}

func (m *model) buildCheckpointsTable(cps []internal.CheckpointInfo) {
	rows := make([]table.Row, len(cps))
	for i := range cps {
		cp := &cps[i]
		rows[i] = table.Row{
			strconv.FormatUint(uint64(cp.Checkpoint), 10),
			strconv.FormatUint(uint64(cp.Version), 10),
			cp.RootID.String(),
			fmt.Sprintf("%d (offset %d, %d-%d)", cp.Leaves.Count, cp.Leaves.StartOffset, cp.Leaves.StartIndex, cp.Leaves.EndIndex),
			fmt.Sprintf("%d (offset %d, %d-%d)", cp.Branches.Count, cp.Branches.StartOffset, cp.Branches.StartIndex, cp.Branches.EndIndex),
		}
	}
	m.table = newTable([]table.Column{
		{Title: "Checkpoint", Width: 10},
		{Title: "Version", Width: 10},
		{Title: "Root", Width: 20},
		{Title: "Leaves", Width: 25},
		{Title: "Branches", Width: 25},
	}, rows, m.tableHeight())
}

func (m *model) buildLeavesTable(leaves []internal.LeafLayout) {
	rows := make([]table.Row, len(leaves))
	for i := range leaves {
		l := &leaves[i]
		rows[i] = table.Row{
			l.ID.String(),
			strconv.FormatUint(uint64(l.Version), 10),
			l.KeyOffset.String(),
			l.ValueOffset.String(),
			hex.EncodeToString(l.Hash[:8]),
		}
	}
	m.table = newTable([]table.Column{
		{Title: "ID", Width: 16},
		{Title: "Version", Width: 10},
		{Title: "KeyOff", Width: 14},
		{Title: "ValOff", Width: 14},
		{Title: "Hash", Width: 18},
	}, rows, m.tableHeight())
}

func (m *model) buildBranchesTable(branches []internal.BranchLayout) {
	rows := make([]table.Row, len(branches))
	for i := range branches {
		b := &branches[i]
		rows[i] = table.Row{
			b.ID.String(),
			strconv.FormatUint(uint64(b.Version), 10),
			strconv.FormatUint(uint64(b.Height), 10),
			b.Size.String(),
			b.Left.String(),
			b.Right.String(),
			hex.EncodeToString(b.Hash[:8]),
		}
	}
	m.table = newTable([]table.Column{
		{Title: "ID", Width: 16},
		{Title: "Version", Width: 10},
		{Title: "Height", Width: 8},
		{Title: "Size", Width: 12},
		{Title: "Left", Width: 16},
		{Title: "Right", Width: 16},
		{Title: "Hash", Width: 18},
	}, rows, m.tableHeight())
}

func (m *model) findCheckpoint(cp uint32) *internal.CheckpointInfo {
	for i := range m.checkpoints {
		if m.checkpoints[i].Checkpoint == cp {
			return &m.checkpoints[i]
		}
	}
	return nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			switch m.view {
			case viewTrees:
				return m, tea.Quit
			case viewChangesets:
				m.view = viewTrees
				m.err = ""
				m.buildTreesTable()
				return m, nil
			case viewCheckpoints:
				m.view = viewChangesets
				m.err = ""
				m.buildChangesetsTable()
				return m, nil
			case viewLeaves, viewBranches:
				m.view = viewCheckpoints
				m.err = ""
				m.buildCheckpointsTable(m.checkpoints)
				return m, nil
			}
		case "enter":
			row := m.table.SelectedRow()
			if row == nil {
				return m, nil
			}
			switch m.view {
			case viewTrees:
				m.selectedTree = row[0]
				m.view = viewChangesets
				m.err = ""
				m.buildChangesetsTable()
				return m, nil
			case viewChangesets:
				m.selectedChangeset = row[0]
				m.view = viewCheckpoints
				m.err = ""
				cps, err := loadCheckpoints(m.dir, m.selectedTree, m.selectedChangeset)
				if err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.checkpoints = cps
				m.buildCheckpointsTable(cps)
				return m, nil
			}
		case "l":
			if m.view == viewCheckpoints {
				row := m.table.SelectedRow()
				if row == nil {
					return m, nil
				}
				cpNum, _ := strconv.ParseUint(row[0], 10, 32)
				cp := m.findCheckpoint(uint32(cpNum))
				if cp == nil {
					return m, nil
				}
				m.selectedCheckpoint = cp.Checkpoint
				leaves, err := loadLeaves(m.dir, m.selectedTree, m.selectedChangeset, cp.Leaves.StartOffset, cp.Leaves.Count)
				if err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.view = viewLeaves
				m.err = ""
				m.buildLeavesTable(leaves)
				return m, nil
			}
		case "b":
			if m.view == viewCheckpoints {
				row := m.table.SelectedRow()
				if row == nil {
					return m, nil
				}
				cpNum, _ := strconv.ParseUint(row[0], 10, 32)
				cp := m.findCheckpoint(uint32(cpNum))
				if cp == nil {
					return m, nil
				}
				m.selectedCheckpoint = cp.Checkpoint
				branches, err := loadBranches(m.dir, m.selectedTree, m.selectedChangeset, cp.Branches.StartOffset, cp.Branches.Count)
				if err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.view = viewBranches
				m.err = ""
				m.buildBranchesTable(branches)
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(m.tableHeight())
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)

	var title string
	switch m.view {
	case viewTrees:
		title = titleStyle.Render("IAVL Trees: " + m.dir)
	case viewChangesets:
		title = titleStyle.Render("Changesets: " + m.selectedTree)
	case viewCheckpoints:
		title = titleStyle.Render(fmt.Sprintf("Checkpoints: %s / %s", m.selectedTree, m.selectedChangeset))
	case viewLeaves:
		title = titleStyle.Render(fmt.Sprintf("Leaves: %s / %s / checkpoint %d", m.selectedTree, m.selectedChangeset, m.selectedCheckpoint))
	case viewBranches:
		title = titleStyle.Render(fmt.Sprintf("Branches: %s / %s / checkpoint %d", m.selectedTree, m.selectedChangeset, m.selectedCheckpoint))
	}

	if m.err != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1)
		return title + "\n" + errStyle.Render("Error: "+m.err) + "\n"
	}

	return title + "\n" + m.table.View() + "\n"
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
