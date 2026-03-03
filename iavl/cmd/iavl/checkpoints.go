package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type checkpointsKeyMap struct {
	Leaves   key.Binding
	Branches key.Binding
	Orphans  key.Binding
}

func (k checkpointsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Leaves, k.Branches, k.Orphans}
}

func (k checkpointsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

type checkpointsView struct {
	dir, treeName, changesetName string
	checkpoints                  []internal.CheckpointInfo
	orphans                      []internal.OrphanEntry
	orphanErr                    string
	orphanMap                    map[internal.NodeID]uint32
	orphanStats                  map[uint32]orphanCounts
	table                        table.Model
	columns                      []table.Column
	keys                         checkpointsKeyMap
	width                        int
	height                       int
}

func newCheckpointsView(
	dir, treeName, changesetName string,
	checkpoints []internal.CheckpointInfo,
	orphans []internal.OrphanEntry,
	orphanErr string,
	orphanMap map[internal.NodeID]uint32,
	orphanStats map[uint32]orphanCounts,
	height int,
) *checkpointsView {
	v := &checkpointsView{
		dir:           dir,
		treeName:      treeName,
		changesetName: changesetName,
		checkpoints:   checkpoints,
		orphans:       orphans,
		orphanErr:     orphanErr,
		orphanMap:     orphanMap,
		orphanStats:   orphanStats,
		keys: checkpointsKeyMap{
			Leaves: key.NewBinding(
				key.WithKeys("l"),
				key.WithHelp("l", "leaves"),
			),
			Branches: key.NewBinding(
				key.WithKeys("b"),
				key.WithHelp("b", "branches"),
			),
			Orphans: key.NewBinding(
				key.WithKeys("o"),
				key.WithHelp("o", "orphans"),
			),
		},
	}
	v.buildTable(height)
	return v
}

func (v *checkpointsView) findCheckpoint(cp uint32) *internal.CheckpointInfo {
	for i := range v.checkpoints {
		if v.checkpoints[i].Checkpoint == cp {
			return &v.checkpoints[i]
		}
	}
	return nil
}

func (v *checkpointsView) buildTable(height int) {
	cps := v.checkpoints
	var totalLeaves, totalBranches, totalLeafOrph, totalBranchOrph int
	rows := make([]table.Row, len(cps))
	for i := range cps {
		cp := &cps[i]
		lc := int(cp.Leaves.Count)
		bc := int(cp.Branches.Count)
		oc := v.orphanStats[cp.Checkpoint]
		totalLeaves += lc
		totalBranches += bc
		totalLeafOrph += oc.leaves
		totalBranchOrph += oc.branches
		orphPct := "-"
		if total := lc + bc; total > 0 {
			orphPct = fmt.Sprintf("%.1f%%", float64(oc.leaves+oc.branches)*100.0/float64(total))
		}
		crcOk := "x"
		if cp.VerifyCRC32() {
			crcOk = "✓"
		}
		rows[i] = table.Row{
			strconv.FormatUint(uint64(cp.Checkpoint), 10),
			strconv.FormatUint(uint64(cp.Version), 10),
			crcOk,
			cp.RootID.String(),
			strconv.Itoa(lc),
			strconv.Itoa(bc),
			strconv.Itoa(oc.leaves),
			strconv.Itoa(oc.branches),
			orphPct,
		}
	}
	totalOrphPct := "-"
	if total := totalLeaves + totalBranches; total > 0 {
		totalOrphPct = fmt.Sprintf("%.1f%%", float64(totalLeafOrph+totalBranchOrph)*100.0/float64(total))
	}
	rows = append(rows, table.Row{
		"━━ TOTAL", "━━", "━━", "━━",
		strconv.Itoa(totalLeaves),
		strconv.Itoa(totalBranches),
		strconv.Itoa(totalLeafOrph),
		strconv.Itoa(totalBranchOrph),
		totalOrphPct,
	})
	cols := []table.Column{
		{Title: "Checkpoint", Width: 10},
		{Title: "Version", Width: 10},
		{Title: "CRC", Width: 5},
		{Title: "Root", Width: 20},
		{Title: "Leaves", Width: 8},
		{Title: "Branches", Width: 10},
		{Title: "LeafOrphans", Width: 15},
		{Title: "BranchOrphans", Width: 15},
		{Title: "Orphan %", Width: 10},
	}
	v.columns = cols
	v.table = newTable(cols, rows, height)
}

func (v *checkpointsView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.table.SetHeight(contentHeight(msg.Height))
		return v, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, v.keys.Leaves):
			row := v.table.SelectedRow()
			if row == nil || isTotalRow(row) {
				return v, nil
			}
			cpNum, _ := strconv.ParseUint(row[0], 10, 32)
			cp := v.findCheckpoint(uint32(cpNum))
			if cp == nil {
				return v, nil
			}
			leaves, err := loadLeaves(v.dir, v.treeName, v.changesetName, cp.Leaves.StartOffset, cp.Leaves.Count)
			if err != nil {
				return v, func() tea.Msg { return errMsg{err} }
			}
			child := newLeavesView(v.dir, v.treeName, v.changesetName, cp.Checkpoint, leaves, v.orphanMap, contentHeight(v.height))
			return v, pushView(child)

		case key.Matches(msg, v.keys.Branches):
			row := v.table.SelectedRow()
			if row == nil || isTotalRow(row) {
				return v, nil
			}
			cpNum, _ := strconv.ParseUint(row[0], 10, 32)
			cp := v.findCheckpoint(uint32(cpNum))
			if cp == nil {
				return v, nil
			}
			branches, err := loadBranches(v.dir, v.treeName, v.changesetName, cp.Branches.StartOffset, cp.Branches.Count)
			if err != nil {
				return v, func() tea.Msg { return errMsg{err} }
			}
			child := newBranchesView(v.dir, v.treeName, v.changesetName, cp.Checkpoint, branches, v.orphanMap, contentHeight(v.height))
			return v, pushView(child)

		case key.Matches(msg, v.keys.Orphans):
			child := newOrphansView(v.dir, v.treeName, v.changesetName, v.orphans, v.orphanErr, contentHeight(v.height))
			return v, pushView(child)
		}
	}
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v *checkpointsView) View() string {
	return v.table.View() + "\n" + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *checkpointsView) Title() string {
	return fmt.Sprintf("Checkpoints: %s / %s", v.treeName, v.changesetName)
}

func (v *checkpointsView) KeyMap() help.KeyMap {
	return v.keys
}

func (v *checkpointsView) HelpDoc() string { return checkpointsHelpDoc }
