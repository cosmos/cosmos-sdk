package main

import (
	"encoding/hex"
	"strconv"

	"github.com/charmbracelet/bubbles/table"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func (m *model) buildBranchesTable(branches []internal.BranchLayout, orphanMap map[internal.NodeID]uint32) {
	rows := make([]table.Row, len(branches))
	for i := range branches {
		b := &branches[i]
		orphStr := "-"
		if v, ok := orphanMap[b.ID]; ok && v != 0 {
			orphStr = strconv.FormatUint(uint64(v), 10)
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
	m.columns = cols
	m.table = newTable(cols, rows, m.tableHeight())
}
