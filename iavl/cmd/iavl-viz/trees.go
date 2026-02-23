package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type treesKeyMap struct {
	Enter      key.Binding
	CommitInfo key.Binding
}

func (k treesKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.CommitInfo}
}

func (k treesKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

type treesView struct {
	dir     string
	table   table.Model
	columns []table.Column
	keys    treesKeyMap
	width   int
	err     string
}

func newTreesView(dir string, height int) *treesView {
	v := &treesView{
		dir: dir,
		keys: treesKeyMap{
			Enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			CommitInfo: key.NewBinding(
				key.WithKeys("c"),
				key.WithHelp("c", "commit info"),
			),
		},
	}
	v.buildTable(height)
	return v
}

func (v *treesView) buildTable(height int) {
	names, err := scanTrees(v.dir)
	if err != nil {
		v.err = err.Error()
		return
	}
	rows := make([]table.Row, len(names))
	for i, name := range names {
		treeDir := filepath.Join(v.dir, "stores", name+".iavl")
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
	cols := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Changesets", Width: 12},
		{Title: "Size", Width: 12},
	}
	v.columns = cols
	v.table = newTable(cols, rows, height)
}

func (v *treesView) Update(msg tea.Msg) (viewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.table.SetHeight(contentHeight(msg.Height))
		return v, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, v.keys.Enter):
			row := v.table.SelectedRow()
			if row == nil {
				return v, nil
			}
			child := newChangesetsView(v.dir, row[0], contentHeight(0))
			return v, pushView(child)
		case key.Matches(msg, v.keys.CommitInfo):
			child := newCommitInfoView(v.dir, contentHeight(0))
			return v, pushView(child)
		}
	}
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v *treesView) View() string {
	if v.err != "" {
		return v.err
	}
	return v.table.View() + "\n" + renderInfoPanel(v.columns, v.table.SelectedRow(), v.width)
}

func (v *treesView) Title() string {
	return "IAVL Trees: " + v.dir
}

func (v *treesView) KeyMap() help.KeyMap {
	return v.keys
}
