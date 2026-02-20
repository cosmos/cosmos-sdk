package main

import (
	"fmt"
	"os"
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
	viewCheckpointOrphans
	viewChangesetOrphans
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
	columns       []table.Column // current table columns (for info view)
	err           string

	dir                string
	selectedTree       string
	selectedChangeset  string
	selectedCheckpoint uint32

	checkpoints        []internal.CheckpointInfo
	orphans            []internal.OrphanEntry
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
		return m.height - 10
	}
	return 18
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

func fmtOrphanPct(orphanSize, leafSize, branchSize int64) string {
	orphanCount := orphanSize / internal.SizeOrphanEntry
	nodeCount := leafSize/internal.SizeLeaf + branchSize/internal.SizeBranch
	if nodeCount == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", float64(orphanCount)/float64(nodeCount)*100)
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

func formatStat(s *runningStats, f func(*runningStats) float64) string {
	if s.n == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f", f(s))
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
			case viewLeaves, viewBranches, viewCheckpointOrphans:
				m.view = viewCheckpoints
				m.err = ""
				m.buildCheckpointsTable(m.checkpoints)
				return m, nil
			case viewChangesetOrphans:
				m.view = viewChangesets
				m.err = ""
				m.buildChangesetsTable()
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
				orphans, orphanErr := loadOrphans(m.dir, m.selectedTree, m.selectedChangeset)
				m.orphans = orphans
				if orphanErr != nil {
					m.err = orphanErr.Error()
				}
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
				m.view = viewCheckpointOrphans
				m.err = ""
				m.buildOrphansTable(m.orphans)
				return m, nil
			}
			if m.view == viewChangesets {
				row := m.table.SelectedRow()
				if row == nil || strings.HasPrefix(row[0], "━━") {
					return m, nil
				}
				m.selectedChangeset = row[0]
				orphans, err := loadOrphans(m.dir, m.selectedTree, m.selectedChangeset)
				if orphans == nil && err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.orphans = orphans
				m.view = viewChangesetOrphans
				if err != nil {
					m.err = err.Error()
				} else {
					m.err = ""
				}
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
		footerText = "enter: checkpoints  o: orphans  w: wal analysis  esc: back  q: quit"
	case viewCheckpoints:
		titleText = fmt.Sprintf("Checkpoints: %s / %s", m.selectedTree, m.selectedChangeset)
		footerText = "l: leaves  b: branches  o: orphans  esc: back  q: quit"
	case viewLeaves:
		titleText = fmt.Sprintf("Leaves: %s / %s / checkpoint %d", m.selectedTree, m.selectedChangeset, m.selectedCheckpoint)
		footerText = "esc: back  q: quit"
	case viewBranches:
		titleText = fmt.Sprintf("Branches: %s / %s / checkpoint %d", m.selectedTree, m.selectedChangeset, m.selectedCheckpoint)
		footerText = "esc: back  q: quit"
	case viewCheckpointOrphans:
		titleText = fmt.Sprintf("Orphans: %s / %s (%d total)", m.selectedTree, m.selectedChangeset, len(m.orphans))
		footerText = "esc: back  q: quit"
	case viewChangesetOrphans:
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

	errBanner := ""
	if m.err != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1)
		// For orphan views, show warning above the table instead of replacing it
		if (m.view == viewChangesetOrphans || m.view == viewCheckpointOrphans) && len(m.orphans) > 0 {
			errBanner = errStyle.Render("Warning: "+m.err) + "\n"
		} else {
			return title + "\n" + errStyle.Render("Error: "+m.err) + "\n" + footer
		}
	}

	extra := ""
	if m.view == viewChangesets && m.sizeBreakdown != "" {
		extra = m.sizeBreakdown
	}

	infoPanel := m.renderInfoPanel()

	return title + "\n" + errBanner + m.table.View() + "\n" + extra + infoPanel + footer
}

func (m model) renderInfoPanel() string {
	row := m.table.SelectedRow()
	if row == nil || len(m.columns) == 0 {
		return ""
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	width := m.width
	if width <= 0 {
		width = 120
	}

	// Build a single truncated line: "Col: val  Col: val  ..."
	sep := "  "
	var line string
	for i, col := range m.columns {
		if i >= len(row) {
			break
		}
		pair := dimStyle.Render(col.Title+":") + " " + row[i]
		candidate := pair
		if line != "" {
			candidate = line + sep + pair
		}
		if lipgloss.Width(candidate) > width {
			break
		}
		line = candidate
	}

	rule := dimStyle.Render(strings.Repeat("─", width))
	return rule + "\n" + line + "\n"
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
