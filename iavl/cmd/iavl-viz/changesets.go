package main

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func (m *model) buildChangesetsTable() {
	treeDir := filepath.Join(m.dir, "stores", m.selectedTree+".iavl")
	entries, err := os.ReadDir(treeDir)
	if err != nil {
		m.err = err.Error()
		return
	}

	// First pass: collect info for each changeset.
	var infos []changesetInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		start, end, compacted, valid := internal.ParseChangesetDirName(e.Name())
		if !valid {
			continue
		}
		csPath := filepath.Join(treeDir, e.Name())
		infos = append(infos, changesetInfo{
			name:       e.Name(),
			start:      start,
			end:        end,
			compacted:  compacted,
			walStart:   loadWALStartVersion(m.dir, m.selectedTree, e.Name()),
			kvSize:     statSize(filepath.Join(csPath, "kv.dat")),
			walSize:    statSize(filepath.Join(csPath, "wal.log")),
			leafSize:   statSize(filepath.Join(csPath, "leaves.dat")),
			branchSize: statSize(filepath.Join(csPath, "branches.dat")),
			cpSize:     statSize(filepath.Join(csPath, "checkpoints.dat")),
			orphanSize: statSize(filepath.Join(csPath, "orphans.dat")),
		})
	}

	// Compute totals.
	var total changesetInfo
	for _, info := range infos {
		total.kvSize += info.kvSize
		total.walSize += info.walSize
		total.leafSize += info.leafSize
		total.branchSize += info.branchSize
		total.cpSize += info.cpSize
		total.orphanSize += info.orphanSize
	}

	// Sort by start version.
	slices.SortFunc(infos, func(a, b changesetInfo) int {
		if a.start < b.start {
			return -1
		}
		if a.start > b.start {
			return 1
		}
		return 0
	})

	// Build rows.
	rows := make([]table.Row, 0, len(infos)+1)
	for _, info := range infos {
		endStr := "-"
		if info.end > 0 {
			endStr = strconv.FormatUint(uint64(info.end), 10)
		}
		compStr := "-"
		if info.compacted > 0 {
			compStr = strconv.FormatUint(uint64(info.compacted), 10)
		}
		rows = append(rows, table.Row{
			info.name,
			strconv.FormatUint(uint64(info.start), 10),
			endStr,
			compStr,
			info.walStart,
			fmtCountAndSize(info.kvSize, 0),
			fmtCountAndSize(info.walSize, 0),
			fmtCountAndSize(info.leafSize, internal.SizeLeaf),
			fmtCountAndSize(info.branchSize, internal.SizeBranch),
			fmtCountAndSize(info.cpSize, internal.CheckpointInfoSize),
			fmtCountAndSize(info.orphanSize, internal.SizeOrphanEntry),
			fmtOrphanPct(info.orphanSize, info.leafSize, info.branchSize),
		})
	}

	// Append TOTAL row.
	rows = append(rows, table.Row{
		"━━ TOTAL ━━", "━━", "━━", "━━", "━━",
		fmtCountAndSize(total.kvSize, 0),
		fmtCountAndSize(total.walSize, 0),
		fmtCountAndSize(total.leafSize, internal.SizeLeaf),
		fmtCountAndSize(total.branchSize, internal.SizeBranch),
		fmtCountAndSize(total.cpSize, internal.CheckpointInfoSize),
		fmtCountAndSize(total.orphanSize, internal.SizeOrphanEntry),
		fmtOrphanPct(total.orphanSize, total.leafSize, total.branchSize),
	})

	// Build size breakdown bar chart.
	breakdown := []sizeEntry{
		{"kv.dat", total.kvSize},
		{"wal.log", total.walSize},
		{"leaves.dat", total.leafSize},
		{"branches", total.branchSize},
		{"checkpts", total.cpSize},
		{"orphans.dat", total.orphanSize},
	}
	m.sizeBreakdown = renderSizeBreakdown(breakdown)

	height := m.tableHeight()
	if m.sizeBreakdown != "" {
		height -= strings.Count(m.sizeBreakdown, "\n")
		if height < 5 {
			height = 5
		}
	}

	cols := []table.Column{
		{Title: "Dir", Width: 20},
		{Title: "Start", Width: 10},
		{Title: "End", Width: 10},
		{Title: "Compacted", Width: 10},
		{Title: "WAL Start", Width: 10},
		{Title: "kv.dat", Width: 8},
		{Title: "wal.log", Width: 8},
		{Title: "Leaves", Width: 18},
		{Title: "Branches", Width: 18},
		{Title: "Checkpts", Width: 14},
		{Title: "Orphans", Width: 18},
		{Title: "Orphan %", Width: 10},
	}
	m.columns = cols
	m.table = newTable(cols, rows, height)
}
