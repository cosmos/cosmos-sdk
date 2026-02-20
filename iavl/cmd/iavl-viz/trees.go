package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/bubbles/table"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

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
	cols := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Changesets", Width: 12},
		{Title: "Size", Width: 12},
	}
	m.columns = cols
	m.table = newTable(cols, rows, m.tableHeight())
}
