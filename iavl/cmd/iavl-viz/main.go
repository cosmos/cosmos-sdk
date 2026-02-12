package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func scanTrees(dir string) ([]string, error) {
	storesDir := filepath.Join(dir, "stores")
	entries, err := os.ReadDir(storesDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".iavl") {
			names = append(names, strings.TrimSuffix(e.Name(), ".iavl"))
		}
	}
	return names, nil
}

func copyMmap[T any](mmap *internal.StructMmap[T], offset, count uint32) []T {
	if count == 0 {
		count = uint32(mmap.Count())
	}
	out := make([]T, count)
	for i := range out {
		out[i] = *mmap.UnsafeItem(offset + uint32(i))
	}
	return out
}

func changesetPath(dir, tree, cs string) string {
	return filepath.Join(dir, "stores", tree+".iavl", cs)
}

func loadCheckpoints(dir, tree, cs string) ([]internal.CheckpointInfo, error) {
	f, err := os.Open(filepath.Join(changesetPath(dir, tree, cs), "checkpoints.dat"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mmap, err := internal.NewStructMmap[internal.CheckpointInfo](f)
	if err != nil {
		return nil, err
	}
	defer mmap.Close()

	return copyMmap(mmap, 0, 0), nil
}

func loadLeaves(dir, tree, cs string, offset, count uint32) ([]internal.LeafLayout, error) {
	f, err := os.Open(filepath.Join(changesetPath(dir, tree, cs), "leaves.dat"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mmap, err := internal.NewNodeReader[internal.LeafLayout](f)
	if err != nil {
		return nil, err
	}
	defer mmap.Close()

	return copyMmap(mmap.StructMmap, offset, count), nil
}

func loadBranches(dir, tree, cs string, offset, count uint32) ([]internal.BranchLayout, error) {
	f, err := os.Open(filepath.Join(changesetPath(dir, tree, cs), "branches.dat"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mmap, err := internal.NewNodeReader[internal.BranchLayout](f)
	if err != nil {
		return nil, err
	}
	defer mmap.Close()

	return copyMmap(mmap.StructMmap, offset, count), nil
}

func loadOrphans(dir, tree, cs string) ([]internal.OrphanLogEntry, error) {
	f, err := os.Open(filepath.Join(changesetPath(dir, tree, cs), "orphans.dat"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rdr, err := internal.ReadOrphanLog(f)
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	var entries []internal.OrphanLogEntry
	for {
		entry, err := rdr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	p := tea.NewProgram(initialModel(dir), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
