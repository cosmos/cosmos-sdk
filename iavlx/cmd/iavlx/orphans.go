package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cosmos/cosmos-sdk/iavlx/internal"
)

type orphansView struct {
	treeName, changesetName string
	orphans                 []internal.OrphanEntry
	warning                 string // non-fatal orphan load warning
	table                   table.Model
	columns                 []table.Column
	width                   int
}

func newOrphansView(dir, treeName, changesetName string, orphans []internal.OrphanEntry, warning string, height int) *orphansView {
	v := &orphansView{
		treeName:      treeName,
		changesetName: changesetName,
		orphans:       orphans,
		warning:       warning,
	}
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
	v.columns = cols
	v.table = newTable(cols, rows, height)
	return v
}

func (v *orphansView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
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

func (v *orphansView) View() string {
	var banner string
	if v.warning != "" && len(v.orphans) > 0 {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1)
		banner = errStyle.Render("Warning: "+v.warning) + "\n"
	}
	return banner + v.table.View() + "\n" + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *orphansView) Title() string {
	return fmt.Sprintf("Orphans: %s / %s (%d total)", v.treeName, v.changesetName, len(v.orphans))
}

func (v *orphansView) KeyMap() help.KeyMap {
	return emptyKeyMap{}
}

func (v *orphansView) HelpDocKey() string { return "orphans.md" }
