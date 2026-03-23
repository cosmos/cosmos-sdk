package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cosmos/cosmos-sdk/iavlx/internal"
)

type changesetsKeyMap struct {
	Enter   key.Binding
	WAL     key.Binding
	Orphans key.Binding
}

func (k changesetsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Orphans, k.WAL}
}

func (k changesetsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

type changesetsView struct {
	dir, treeName string
	table         table.Model
	columns       []table.Column
	keys          changesetsKeyMap
	sizeBreakdown string
	width         int
	height        int
}

func newChangesetsView(dir, treeName string, height int) *changesetsView {
	v := &changesetsView{
		dir:      dir,
		treeName: treeName,
		keys: changesetsKeyMap{
			Enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "checkpoints"),
			),
			WAL: key.NewBinding(
				key.WithKeys("w"),
				key.WithHelp("w", "wal analysis"),
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

func (v *changesetsView) buildTable(height int) {
	treeDir := filepath.Join(v.dir, "stores", v.treeName+".iavl")
	entries, err := os.ReadDir(treeDir)
	if err != nil {
		return
	}

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
			walStart:   loadWALStartVersion(v.dir, v.treeName, e.Name()),
			kvSize:     statSize(filepath.Join(csPath, "kv.dat")),
			walSize:    statSize(filepath.Join(csPath, "wal.log")),
			leafSize:   statSize(filepath.Join(csPath, "leaves.dat")),
			branchSize: statSize(filepath.Join(csPath, "branches.dat")),
			cpSize:     statSize(filepath.Join(csPath, "checkpoints.dat")),
			orphanSize: statSize(filepath.Join(csPath, "orphans.dat")),
		})
	}

	var total changesetInfo
	for _, info := range infos {
		total.kvSize += info.kvSize
		total.walSize += info.walSize
		total.leafSize += info.leafSize
		total.branchSize += info.branchSize
		total.cpSize += info.cpSize
		total.orphanSize += info.orphanSize
	}

	slices.SortFunc(infos, func(a, b changesetInfo) int {
		if a.start < b.start {
			return -1
		}
		if a.start > b.start {
			return 1
		}
		return 0
	})

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

	breakdown := []sizeEntry{
		{"kv.dat", total.kvSize},
		{"wal.log", total.walSize},
		{"leaves.dat", total.leafSize},
		{"branches", total.branchSize},
		{"checkpts", total.cpSize},
		{"orphans.dat", total.orphanSize},
	}
	v.sizeBreakdown = renderSizeBreakdown(breakdown)

	if v.sizeBreakdown != "" {
		height -= strings.Count(v.sizeBreakdown, "\n")
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
	v.columns = cols
	v.table = newTable(cols, rows, height)
}

func (v *changesetsView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		h := contentHeight(msg.Height)
		if v.sizeBreakdown != "" {
			h -= strings.Count(v.sizeBreakdown, "\n")
			if h < 5 {
				h = 5
			}
		}
		v.table.SetHeight(h)
		return v, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, v.keys.Enter):
			row := v.table.SelectedRow()
			if row == nil || isTotalRow(row) {
				return v, nil
			}
			csName := row[0]
			cps, err := loadCheckpoints(v.dir, v.treeName, csName)
			if err != nil {
				return v, func() tea.Msg { return errMsg{err} }
			}
			orphans, orphanErr := loadOrphans(v.dir, v.treeName, csName)
			var orphanErrStr string
			if orphanErr != nil {
				orphanErrStr = orphanErr.Error()
			}
			orphanMap := make(map[internal.NodeID]uint32, len(orphans))
			orphanStats := make(map[uint32]orphanCounts)
			for _, o := range orphans {
				orphanMap[o.NodeID] = o.OrphanedVersion
				c := orphanStats[o.NodeID.Checkpoint()]
				if o.NodeID.IsLeaf() {
					c.leaves++
				} else {
					c.branches++
				}
				orphanStats[o.NodeID.Checkpoint()] = c
			}
			child := newCheckpointsView(v.dir, v.treeName, csName, cps, orphans, orphanErrStr, orphanMap, orphanStats, contentHeight(v.height))
			return v, pushView(child)

		case key.Matches(msg, v.keys.WAL):
			row := v.table.SelectedRow()
			if row == nil || isTotalRow(row) {
				return v, nil
			}
			csName := row[0]
			info, total, err := loadWALAnalysis(v.dir, v.treeName, csName)
			if err != nil {
				return v, func() tea.Msg { return errMsg{err} }
			}
			walSz := walFileSize(v.dir, v.treeName, csName)
			child := newWALAnalysisView(v.dir, v.treeName, csName, info, total, walSz, contentHeight(v.height))
			return v, pushView(child)

		case key.Matches(msg, v.keys.Orphans):
			row := v.table.SelectedRow()
			if row == nil || isTotalRow(row) {
				return v, nil
			}
			csName := row[0]
			orphans, err := loadOrphans(v.dir, v.treeName, csName)
			if orphans == nil && err != nil {
				return v, func() tea.Msg { return errMsg{err} }
			}
			var warning string
			if err != nil {
				warning = err.Error()
			}
			child := newOrphansView(v.dir, v.treeName, csName, orphans, warning, contentHeight(v.height))
			return v, pushView(child)
		}
	}
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v *changesetsView) View() string {
	extra := ""
	if v.sizeBreakdown != "" {
		extra = v.sizeBreakdown
	}
	return v.table.View() + "\n" + extra + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *changesetsView) Title() string {
	return fmt.Sprintf("Changesets: %s", v.treeName)
}

func (v *changesetsView) KeyMap() help.KeyMap {
	return v.keys
}

func (v *changesetsView) HelpDocKey() string { return "changeset.md" }

func (v *changesetsView) refresh() {
	fresh := newChangesetsView(v.dir, v.treeName, contentHeight(v.height))
	fresh.width, fresh.height = v.width, v.height
	*v = *fresh
}
