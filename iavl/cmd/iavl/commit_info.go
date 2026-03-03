package main

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type commitInfoView struct {
	dir     string
	table   table.Model
	columns []table.Column
	width   int
}

func newCommitInfoView(dir string, height int) *commitInfoView {
	v := &commitInfoView{dir: dir}
	infos := loadAllCommitInfos(dir)
	rows := make([]table.Row, len(infos))
	for i, ci := range infos {
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
	cols := []table.Column{
		{Title: "Version", Width: 10},
		{Title: "Hash", Width: 66},
		{Title: "Stores", Width: 50},
		{Title: "Error", Width: 30},
	}
	v.columns = cols
	v.table = newTable(cols, rows, height)
	return v
}

func (v *commitInfoView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
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

func (v *commitInfoView) View() string {
	return v.table.View() + "\n" + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *commitInfoView) Title() string {
	return "Commit Info: " + v.dir
}

func (v *commitInfoView) KeyMap() help.KeyMap {
	return emptyKeyMap{}
}

func (v *commitInfoView) HelpDoc() string { return commitInfoHelpDoc }
