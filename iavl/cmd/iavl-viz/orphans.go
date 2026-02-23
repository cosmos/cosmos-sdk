package main

import (
	"strconv"

	"github.com/charmbracelet/bubbles/table"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func (m *model) buildOrphansTable(orphans []internal.OrphanEntry) {
	rows := make([]table.Row, len(orphans))
	for i := range orphans {
		o := &orphans[i]
		nodeType := "branch"
		if o.NodeID.IsLeaf() {
			nodeType = "leaf"
		}
		rows[i] = table.Row{
			o.NodeID.String(),
			nodeType,
			strconv.FormatUint(uint64(o.NodeID.Checkpoint()), 10),
			strconv.FormatUint(uint64(o.OrphanedVersion), 10),
		}
	}
	cols := []table.Column{
		{Title: "NodeID", Width: 16},
		{Title: "Type", Width: 8},
		{Title: "Checkpoint", Width: 12},
		{Title: "OrphanedVer", Width: 12},
	}
	m.columns = cols
	m.table = newTable(cols, rows, m.tableHeight())
}
