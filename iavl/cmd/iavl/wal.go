package main

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// walAnalysisView shows per-version WAL summary.

type walAnalysisKeyMap struct {
	Enter key.Binding
}

func (k walAnalysisKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter}
}

func (k walAnalysisKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

type walAnalysisView struct {
	treeName, changesetName string
	walAnalysis             []walVersionInfo
	walSize                 string
	table                   table.Model
	columns                 []table.Column
	keys                    walAnalysisKeyMap
	width                   int
}

func newWALAnalysisView(dir, treeName, changesetName string, info []walVersionInfo, total walVersionInfo, walSize string, height int) *walAnalysisView {
	v := &walAnalysisView{
		treeName:      treeName,
		changesetName: changesetName,
		walAnalysis:   info,
		walSize:       walSize,
		keys: walAnalysisKeyMap{
			Enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "entries"),
			),
		},
	}

	rows := make([]table.Row, 0, len(info)+1)
	for i := range info {
		vi := &info[i]
		rows = append(rows, table.Row{
			strconv.FormatUint(vi.version, 10),
			strconv.Itoa(vi.sets),
			strconv.Itoa(vi.deletes),
			formatStat(&vi.keyStats, (*runningStats).avg),
			formatStat(&vi.keyStats, (*runningStats).stddev),
			formatStat(&vi.valStats, (*runningStats).avg),
			formatStat(&vi.valStats, (*runningStats).stddev),
			strconv.Itoa(vi.offset),
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
	cols := []table.Column{
		{Title: "Version", Width: 10},
		{Title: "Sets", Width: 8},
		{Title: "Deletes", Width: 8},
		{Title: "Avg Key", Width: 10},
		{Title: "Key StdDev", Width: 10},
		{Title: "Avg Val", Width: 10},
		{Title: "Val StdDev", Width: 10},
		{Title: "Offset", Width: 10},
	}
	v.columns = cols
	v.table = newTable(cols, rows, height)
	return v
}

func (v *walAnalysisView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.table.SetHeight(contentHeight(msg.Height))
		return v, nil
	case tea.KeyMsg:
		if key.Matches(msg, v.keys.Enter) {
			row := v.table.SelectedRow()
			if row == nil || isTotalRow(row) {
				return v, nil
			}
			ver, _ := strconv.ParseUint(row[0], 10, 64)
			for _, info := range v.walAnalysis {
				if info.version == ver {
					child := newWALEntriesView(v.treeName, v.changesetName, ver, info.entries, contentHeight(0))
					return v, pushView(child)
				}
			}
			return v, nil
		}
	}
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v *walAnalysisView) View() string {
	return v.table.View() + "\n" + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *walAnalysisView) Title() string {
	return fmt.Sprintf("WAL Analysis: %s / %s (wal.log: %s)", v.treeName, v.changesetName, v.walSize)
}

func (v *walAnalysisView) KeyMap() help.KeyMap {
	return v.keys
}

func (v *walAnalysisView) HelpDoc() string { return walAnalysisHelpDoc }

// walEntriesView shows individual WAL entries for a version.

type walEntriesView struct {
	treeName, changesetName string
	version                 uint64
	table                   table.Model
	columns                 []table.Column
	width                   int
}

func newWALEntriesView(treeName, changesetName string, version uint64, entries []walEntry, height int) *walEntriesView {
	v := &walEntriesView{
		treeName:      treeName,
		changesetName: changesetName,
		version:       version,
	}
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
	cols := []table.Column{
		{Title: "Key", Width: 50},
		{Title: "Value", Width: 50},
		{Title: "Delete", Width: 8},
	}
	v.columns = cols
	v.table = newTable(cols, rows, height)
	return v
}

func (v *walEntriesView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
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

func (v *walEntriesView) View() string {
	return v.table.View() + "\n" + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *walEntriesView) Title() string {
	return fmt.Sprintf("WAL Entries: %s / %s / version %d", v.treeName, v.changesetName, v.version)
}

func (v *walEntriesView) KeyMap() help.KeyMap {
	return emptyKeyMap{}
}

func (v *walEntriesView) HelpDoc() string { return walEntriesHelpDoc }
