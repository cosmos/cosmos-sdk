package main

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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
	viewOrphans
	viewWALAnalysis
	viewWALEntries
	viewCommitInfo
)

type orphanCounts struct {
	leaves   int
	branches int
}

type model struct {
	view          view
	width, height int
	table         table.Model
	err           string

	dir                string
	selectedTree       string
	selectedChangeset  string
	selectedCheckpoint uint32

	checkpoints        []internal.CheckpointInfo
	orphans            []internal.OrphanLogEntry
	orphanMap          map[internal.NodeID]uint32 // NodeID → OrphanedVersion
	orphanStats        map[uint32]orphanCounts    // checkpoint → {leaf orphan count, branch orphan count}
	walAnalysis        []walVersionInfo
	walTotal           walVersionInfo
	selectedWALVersion uint64
	walSize            string
	sizeBreakdown      string
	commitInfos        []commitInfoResult
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
		return m.height - 8
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

func statSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func fmtCountAndSize(size, structSize int64) string {
	if size == 0 {
		return "-"
	}
	if structSize > 0 {
		return fmt.Sprintf("%d (%s)", size/structSize, humanSize(size))
	}
	return humanSize(size)
}

type sizeEntry struct {
	label string
	size  int64
}

func renderSizeBreakdown(sizes []sizeEntry) string {
	var total int64
	for _, e := range sizes {
		total += e.size
	}
	if total == 0 {
		return ""
	}

	sort.Slice(sizes, func(i, j int) bool {
		return sizes[i].size > sizes[j].size
	})

	const barWidth = 30
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	var b strings.Builder
	b.WriteString(dimStyle.Render(fmt.Sprintf("  Size Breakdown (%s total)", humanSize(total))))
	b.WriteByte('\n')

	for _, e := range sizes {
		if e.size == 0 {
			continue
		}
		pct := float64(e.size) * 100.0 / float64(total)
		filled := float64(barWidth) * float64(e.size) / float64(total)
		fullBlocks := int(filled)
		frac := filled - float64(fullBlocks)

		bar := strings.Repeat("█", fullBlocks)
		if frac > 0.0625 {
			fractional := []rune{'▏', '▎', '▍', '▌', '▋', '▊', '▉'}
			idx := int(frac * 8)
			if idx > 6 {
				idx = 6
			}
			bar += string(fractional[idx])
		}
		if bar == "" {
			bar = "▏"
		}

		b.WriteString(fmt.Sprintf("  %-12s %s %5.1f%%  %s\n",
			e.label,
			barStyle.Render(fmt.Sprintf("%-*s", barWidth, bar)),
			pct,
			humanSize(e.size),
		))
	}
	return b.String()
}

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
			fmtCountAndSize(info.orphanSize, 0),
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
		fmtCountAndSize(total.orphanSize, 0),
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
		height -= strings.Count(m.sizeBreakdown, "\n") + 1
		if height < 5 {
			height = 5
		}
	}

	m.table = newTable([]table.Column{
		{Title: "Dir", Width: 20},
		{Title: "Start", Width: 10},
		{Title: "End", Width: 10},
		{Title: "Compacted", Width: 10},
		{Title: "WAL Start", Width: 10},
		{Title: "kv.dat", Width: 10},
		{Title: "wal.log", Width: 10},
		{Title: "Leaves", Width: 14},
		{Title: "Branches", Width: 14},
		{Title: "Checkpts", Width: 14},
		{Title: "orphans.dat", Width: 12},
	}, rows, height)
}

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
		rows[i] = table.Row{
			strconv.FormatUint(uint64(cp.Checkpoint), 10),
			strconv.FormatUint(uint64(cp.Version), 10),
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
		"━━ TOTAL", "━━", "━━",
		strconv.Itoa(totalLeaves),
		strconv.Itoa(totalBranches),
		strconv.Itoa(totalLeafOrph),
		strconv.Itoa(totalBranchOrph),
		totalOrphPct,
	})
	m.table = newTable([]table.Column{
		{Title: "Checkpoint", Width: 10},
		{Title: "Version", Width: 10},
		{Title: "Root", Width: 20},
		{Title: "Leaves", Width: 8},
		{Title: "Branches", Width: 10},
		{Title: "LeafOrphans", Width: 15},
		{Title: "BranchOrphans", Width: 15},
		{Title: "Orphan %", Width: 10},
	}, rows, m.tableHeight())
}

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
	m.table = newTable([]table.Column{
		{Title: "ID", Width: 16},
		{Title: "Version", Width: 10},
		{Title: "KeyOff", Width: 14},
		{Title: "ValOff", Width: 14},
		{Title: "Orphaned", Width: 10},
		{Title: "Hash", Width: 66},
	}, rows, m.tableHeight())
}

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
	m.table = newTable([]table.Column{
		{Title: "ID", Width: 16},
		{Title: "Version", Width: 10},
		{Title: "Height", Width: 8},
		{Title: "Size", Width: 12},
		{Title: "Left", Width: 16},
		{Title: "Right", Width: 16},
		{Title: "Orphaned", Width: 10},
		{Title: "Hash", Width: 66},
	}, rows, m.tableHeight())
}

func (m *model) buildOrphansTable(orphans []internal.OrphanLogEntry) {
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
	m.table = newTable([]table.Column{
		{Title: "NodeID", Width: 16},
		{Title: "Type", Width: 8},
		{Title: "Checkpoint", Width: 12},
		{Title: "OrphanedVer", Width: 12},
	}, rows, m.tableHeight())
}

func formatStat(s *runningStats, f func(*runningStats) float64) string {
	if s.n == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f", f(s))
}

func (m *model) buildWALAnalysisTable(info []walVersionInfo, total walVersionInfo) {
	rows := make([]table.Row, 0, len(info)+1)
	for i := range info {
		v := &info[i]
		rows = append(rows, table.Row{
			strconv.FormatUint(v.version, 10),
			strconv.Itoa(v.sets),
			strconv.Itoa(v.deletes),
			formatStat(&v.keyStats, (*runningStats).avg),
			formatStat(&v.keyStats, (*runningStats).stddev),
			formatStat(&v.valStats, (*runningStats).avg),
			formatStat(&v.valStats, (*runningStats).stddev),
			strconv.Itoa(v.offset),
		})
	}
	rows = append(rows, table.Row{
		"━━ TOTAL",
		strconv.Itoa(total.sets),
		strconv.Itoa(total.deletes),
		formatStat(&total.keyStats, (*runningStats).avg),
		formatStat(&total.keyStats, (*runningStats).stddev),
		formatStat(&total.valStats, (*runningStats).avg),
		formatStat(&total.valStats, (*runningStats).stddev),
		"━━",
	})
	m.table = newTable([]table.Column{
		{Title: "Version", Width: 10},
		{Title: "Sets", Width: 8},
		{Title: "Deletes", Width: 8},
		{Title: "Avg Key", Width: 10},
		{Title: "Key StdDev", Width: 10},
		{Title: "Avg Val", Width: 10},
		{Title: "Val StdDev", Width: 10},
		{Title: "Offset", Width: 10},
	}, rows, m.tableHeight())
}

func (m *model) buildWALEntriesTable(entries []walEntry) {
	rows := make([]table.Row, len(entries))
	for i := range entries {
		e := &entries[i]
		del := "no"
		if e.delete {
			del = "yes"
		}
		rows[i] = table.Row{
			hex.EncodeToString(e.key),
			hex.EncodeToString(e.value),
			del,
		}
	}
	m.table = newTable([]table.Column{
		{Title: "Key", Width: 50},
		{Title: "Value", Width: 50},
		{Title: "Delete", Width: 8},
	}, rows, m.tableHeight())
}

func (m *model) buildCommitInfoTable() {
	rows := make([]table.Row, len(m.commitInfos))
	for i, ci := range m.commitInfos {
		if ci.err != nil {
			rows[i] = table.Row{ci.version, "-", "-", ci.err.Error()}
		} else {
			var storeNames []string
			for _, si := range ci.info.StoreInfos {
				storeNames = append(storeNames, si.Name)
			}
			rows[i] = table.Row{
				strconv.FormatInt(ci.info.Version, 10),
				hex.EncodeToString(ci.info.Hash()),
				strings.Join(storeNames, ", "),
				"-",
			}
		}
	}
	m.table = newTable([]table.Column{
		{Title: "Version", Width: 10},
		{Title: "Hash", Width: 66},
		{Title: "Stores", Width: 50},
		{Title: "Error", Width: 30},
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
		case "q":
			return m, tea.Quit
		case "esc":
			switch m.view {
			case viewCommitInfo:
				m.view = viewTrees
				m.err = ""
				m.buildTreesTable()
				return m, nil
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
			case viewWALAnalysis:
				m.view = viewChangesets
				m.err = ""
				m.buildChangesetsTable()
				return m, nil
			case viewWALEntries:
				m.view = viewWALAnalysis
				m.err = ""
				m.buildWALAnalysisTable(m.walAnalysis, m.walTotal)
				return m, nil
			case viewLeaves, viewBranches, viewOrphans:
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
				if strings.HasPrefix(row[0], "━━") {
					return m, nil
				}
				m.selectedChangeset = row[0]
				m.view = viewCheckpoints
				m.err = ""
				cps, err := loadCheckpoints(m.dir, m.selectedTree, m.selectedChangeset)
				if err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.checkpoints = cps
				orphans, _ := loadOrphans(m.dir, m.selectedTree, m.selectedChangeset)
				m.orphans = orphans
				m.orphanMap = make(map[internal.NodeID]uint32, len(orphans))
				m.orphanStats = make(map[uint32]orphanCounts)
				for _, o := range orphans {
					m.orphanMap[o.NodeID] = o.OrphanedVersion
					c := m.orphanStats[o.NodeID.Checkpoint()]
					if o.NodeID.IsLeaf() {
						c.leaves++
					} else {
						c.branches++
					}
					m.orphanStats[o.NodeID.Checkpoint()] = c
				}
				m.buildCheckpointsTable(cps)
				return m, nil
			case viewWALAnalysis:
				if strings.HasPrefix(row[0], "━━") {
					return m, nil
				}
				ver, _ := strconv.ParseUint(row[0], 10, 64)
				for _, info := range m.walAnalysis {
					if info.version == ver {
						m.selectedWALVersion = ver
						m.view = viewWALEntries
						m.err = ""
						m.buildWALEntriesTable(info.entries)
						return m, nil
					}
				}
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
				m.buildLeavesTable(leaves, m.orphanMap)
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
				m.buildBranchesTable(branches, m.orphanMap)
				return m, nil
			}
		case "o":
			if m.view == viewCheckpoints {
				m.view = viewOrphans
				m.err = ""
				m.buildOrphansTable(m.orphans)
				return m, nil
			}
		case "w":
			if m.view == viewChangesets {
				row := m.table.SelectedRow()
				if row == nil || strings.HasPrefix(row[0], "━━") {
					return m, nil
				}
				m.selectedChangeset = row[0]
				info, total, err := loadWALAnalysis(m.dir, m.selectedTree, m.selectedChangeset)
				if err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.walAnalysis = info
				m.walTotal = total
				m.walSize = walFileSize(m.dir, m.selectedTree, m.selectedChangeset)
				m.view = viewWALAnalysis
				m.err = ""
				m.buildWALAnalysisTable(info, total)
				return m, nil
			}
		case "c":
			if m.view == viewTrees {
				m.commitInfos = loadAllCommitInfos(m.dir)
				m.view = viewCommitInfo
				m.err = ""
				m.buildCommitInfoTable()
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
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	var titleText string
	var footerText string
	switch m.view {
	case viewTrees:
		titleText = "IAVL Trees: " + m.dir
		footerText = "enter: select  c: commit info  q: quit"
	case viewChangesets:
		titleText = "Changesets: " + m.selectedTree
		footerText = "enter: checkpoints  w: wal analysis  esc: back  q: quit"
	case viewCheckpoints:
		titleText = fmt.Sprintf("Checkpoints: %s / %s", m.selectedTree, m.selectedChangeset)
		footerText = "l: leaves  b: branches  o: orphans  esc: back  q: quit"
	case viewLeaves:
		titleText = fmt.Sprintf("Leaves: %s / %s / checkpoint %d", m.selectedTree, m.selectedChangeset, m.selectedCheckpoint)
		footerText = "esc: back  q: quit"
	case viewBranches:
		titleText = fmt.Sprintf("Branches: %s / %s / checkpoint %d", m.selectedTree, m.selectedChangeset, m.selectedCheckpoint)
		footerText = "esc: back  q: quit"
	case viewOrphans:
		titleText = fmt.Sprintf("Orphans: %s / %s (%d total)", m.selectedTree, m.selectedChangeset, len(m.orphans))
		footerText = "esc: back  q: quit"
	case viewWALAnalysis:
		titleText = fmt.Sprintf("WAL Analysis: %s / %s (wal.log: %s)", m.selectedTree, m.selectedChangeset, m.walSize)
		footerText = "enter: entries  esc: back  q: quit"
	case viewWALEntries:
		titleText = fmt.Sprintf("WAL Entries: %s / %s / version %d", m.selectedTree, m.selectedChangeset, m.selectedWALVersion)
		footerText = "esc: back  q: quit"
	case viewCommitInfo:
		titleText = "Commit Info: " + m.dir
		footerText = "esc: back  q: quit"
	}

	title := boxStyle.Render(titleText)
	footer := boxStyle.Render(footerText)

	if m.err != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1)
		return title + "\n" + errStyle.Render("Error: "+m.err) + "\n" + footer
	}

	extra := ""
	if m.view == viewChangesets && m.sizeBreakdown != "" {
		extra = m.sizeBreakdown
	}

	return title + "\n" + m.table.View() + "\n" + extra + footer
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
