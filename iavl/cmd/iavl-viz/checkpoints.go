package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func (m *model) buildCheckpointsTable(cps []internal.CheckpointInfo) {
	var totalLeaves, totalBranches, totalLeafOrph, totalBranchOrph int
	rows := make([]table.Row, len(cps))
	for i := range cps {
		cp := &cps[i]
		lc := int(cp.Leaves.Count)
		bc := int(cp.Branches.Count)
		oc := m.orphanStats[cp.Checkpoint]
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
	m.columns = cols
	m.table = newTable(cols, rows, m.tableHeight())
}
