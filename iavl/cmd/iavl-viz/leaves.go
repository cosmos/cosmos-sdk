package main

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type leavesView struct {
	treeName, changesetName string
	checkpoint              uint32
	table                   table.Model
	columns                 []table.Column
	width                   int
}

func newLeavesView(dir, treeName, changesetName string, checkpoint uint32, leaves []internal.LeafLayout, orphanMap map[internal.NodeID]uint32, height int) *leavesView {
	v := &leavesView{
		treeName:      treeName,
		changesetName: changesetName,
		checkpoint:    checkpoint,
	}
	rows := make([]table.Row, len(leaves))
	for i := range leaves {
		l := &leaves[i]
		orphStr := "-"
		if ver, ok := orphanMap[l.ID]; ok && ver != 0 {
			orphStr = strconv.FormatUint(uint64(ver), 10)
		}
		rows[i] = table.Row{
			l.ID.String(),
			strconv.FormatUint(uint64(l.Version), 10),
			l.KeyOffset.String(),
			l.ValueOffset.String(),
			orphStr,
			hex.EncodeToString(l.Hash[:]),
		}
	}
	cols := []table.Column{
		{Title: "ID", Width: 16},
		{Title: "Version", Width: 10},
		{Title: "KeyOff", Width: 14},
		{Title: "ValOff", Width: 14},
		{Title: "Orphaned", Width: 10},
		{Title: "Hash", Width: 66},
	}
	v.columns = cols
	v.table = newTable(cols, rows, height)
	return v
}

func (v *leavesView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
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

func (v *leavesView) View() string {
	return v.table.View() + "\n" + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *leavesView) Title() string {
	return fmt.Sprintf("Leaves: %s / %s / checkpoint %d", v.treeName, v.changesetName, v.checkpoint)
}

func (v *leavesView) KeyMap() help.KeyMap {
	return emptyKeyMap{}
}
