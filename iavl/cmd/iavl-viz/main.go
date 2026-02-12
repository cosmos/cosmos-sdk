package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type changesetEntry struct {
	dirName   string
	start     uint32
	end       uint32
	compacted uint32
	size      int64
}

type treeEntry struct {
	name       string
	path       string
	changesets []changesetEntry
	totalSize  int64
}

func scanIAVLDirs(dir string) ([]treeEntry, error) {
	dir = filepath.Join(dir, "stores") // stores/ dir contains the iavl tree dirs
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var trees []treeEntry
	for _, e := range entries {
		if !e.IsDir() || !strings.HasSuffix(e.Name(), ".iavl") {
			continue
		}
		treePath := filepath.Join(dir, e.Name())
		te := treeEntry{
			name: strings.TrimSuffix(e.Name(), ".iavl"),
			path: treePath,
		}

		subEntries, err := os.ReadDir(treePath)
		if err != nil {
			continue
		}
		for _, se := range subEntries {
			if !se.IsDir() {
				continue
			}
			start, end, compacted, valid := internal.ParseChangesetDirName(se.Name())
			if !valid {
				continue
			}
			ce := changesetEntry{
				dirName:   se.Name(),
				start:     start,
				end:       end,
				compacted: compacted,
			}
			csPath := filepath.Join(treePath, se.Name())
			_ = filepath.WalkDir(csPath, func(_ string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				if info, err := d.Info(); err == nil {
					ce.size += info.Size()
				}
				return nil
			})
			te.totalSize += ce.size
			te.changesets = append(te.changesets, ce)
		}
		trees = append(trees, te)
	}
	return trees, nil
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	trees, err := scanIAVLDirs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error scanning directory: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(dir, trees), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
