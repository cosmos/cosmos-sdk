package main

import (
	"encoding/hex"
	"strconv"

	"github.com/charmbracelet/bubbles/table"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func (m *model) buildLeavesTable(leaves []internal.LeafLayout, orphanMap map[internal.NodeID]uint32) {
	rows := make([]table.Row, len(leaves))
	for i := range leaves {
		l := &leaves[i]
		orphStr := "-"
		if v, ok := orphanMap[l.ID]; ok && v != 0 {
			orphStr = strconv.FormatUint(uint64(v), 10)
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
	m.columns = cols
	m.table = newTable(cols, rows, m.tableHeight())
}
