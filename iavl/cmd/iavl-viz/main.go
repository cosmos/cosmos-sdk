package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type runningStats struct {
	n    int
	mean float64
	m2   float64
}

func (s *runningStats) add(x float64) {
	s.n++
	delta := x - s.mean
	s.mean += delta / float64(s.n)
	s.m2 += delta * (x - s.mean)
}

func (s *runningStats) avg() float64 {
	return s.mean
}

func (s *runningStats) stddev() float64 {
	if s.n > 0 {
		return math.Sqrt(s.m2 / float64(s.n))
	}
	return 0
}

type walEntry struct {
	key    []byte
	value  []byte
	delete bool
}

type walVersionInfo struct {
	version  uint64
	offset   int
	sets     int
	deletes  int
	keyStats runningStats
	valStats runningStats
	entries  []walEntry
}

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

func loadWALStartVersion(dir, tree, cs string) string {
	f, err := os.Open(filepath.Join(changesetPath(dir, tree, cs), "wal.log"))
	if err != nil {
		return "-"
	}
	defer f.Close()

	for entry, err := range internal.ReadWAL(f) {
		if err != nil {
			return "-"
		}
		return strconv.FormatUint(entry.Version, 10)
	}
	return "-"
}

func loadWALAnalysis(dir, tree, cs string) ([]walVersionInfo, walVersionInfo, error) {
	f, err := os.Open(filepath.Join(changesetPath(dir, tree, cs), "wal.log"))
	if err != nil {
		return nil, walVersionInfo{}, err
	}
	defer f.Close()

	var results []walVersionInfo
	var cur walVersionInfo
	var total walVersionInfo
	started := false

	for entry, err := range internal.ReadWAL(f) {
		if err != nil {
			return nil, walVersionInfo{}, err
		}

		if !started || entry.Version != cur.version {
			if started {
				results = append(results, cur)
			}
			cur = walVersionInfo{
				version: entry.Version,
				offset:  entry.KeyOffset,
			}
			started = true
		}

		switch entry.Op {
		case internal.WALOpSet:
			cur.sets++
			total.sets++
			keyLen := float64(len(entry.Key.UnsafeBytes()))
			valLen := float64(len(entry.Value.UnsafeBytes()))
			cur.keyStats.add(keyLen)
			cur.valStats.add(valLen)
			total.keyStats.add(keyLen)
			total.valStats.add(valLen)
			cur.entries = append(cur.entries, walEntry{key: entry.Key.SafeCopy(), value: entry.Value.SafeCopy(), delete: false})
		case internal.WALOpDelete:
			cur.deletes++
			total.deletes++
			keyLen := float64(len(entry.Key.UnsafeBytes()))
			cur.keyStats.add(keyLen)
			total.keyStats.add(keyLen)
			cur.entries = append(cur.entries, walEntry{key: entry.Key.SafeCopy(), value: nil, delete: true})
		case internal.WALOpCommit:
			results = append(results, cur)
			cur = walVersionInfo{}
			started = false
		}
	}

	if started {
		results = append(results, cur)
	}

	return results, total, nil
}

func walFileSize(dir, tree, cs string) string {
	info, err := os.Stat(filepath.Join(changesetPath(dir, tree, cs), "wal.log"))
	if err != nil {
		return "?"
	}
	return humanSize(info.Size())
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
