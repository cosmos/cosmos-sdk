package main

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cosmos/cosmos-sdk/iavlx/internal"
)

type branchesView struct {
	treeName, changesetName string
	checkpoint              uint32
	table                   table.Model
	columns                 []table.Column
	width                   int
}

func newBranchesView(dir, treeName, changesetName string, checkpoint uint32, branches []internal.BranchLayout, orphanMap map[internal.NodeID]uint32, height int) *branchesView {
	v := &branchesView{
		treeName:      treeName,
		changesetName: changesetName,
		checkpoint:    checkpoint,
	}
	rows := make([]table.Row, len(branches))
	for i := range branches {
		b := &branches[i]
		orphStr := "-"
		if ver, ok := orphanMap[b.ID]; ok && ver != 0 {
			orphStr = strconv.FormatUint(uint64(ver), 10)
		}
		rows[i] = table.Row{
			b.ID.String(),
			strconv.FormatUint(uint64(b.Version), 10),
			strconv.FormatUint(uint64(b.Height), 10),
			b.Size.String(),
			b.Left.String(),
			b.Right.String(),
			orphStr,
			hex.EncodeToString(b.Hash[:]),
		}
	}
	cols := []table.Column{
		{Title: "ID", Width: 16},
		{Title: "Version", Width: 10},
		{Title: "Height", Width: 8},
		{Title: "Size", Width: 12},
		{Title: "Left", Width: 16},
		{Title: "Right", Width: 16},
		{Title: "Orphaned", Width: 10},
		{Title: "Hash", Width: 66},
	}
	v.columns = cols
	v.table = newTable(cols, rows, height)
	return v
}

func (v *branchesView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.table.SetHeight(contentHeight(msg.Height))
		return v, nil
	}
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v *branchesView) View() string {
	return v.table.View() + "\n" + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *branchesView) Title() string {
	return fmt.Sprintf("Branches: %s / %s / checkpoint %d", v.treeName, v.changesetName, v.checkpoint)
}

func (v *branchesView) KeyMap() help.KeyMap {
	return emptyKeyMap{}
}

func (v *branchesView) HelpDocKey() string { return "branches.md" }
