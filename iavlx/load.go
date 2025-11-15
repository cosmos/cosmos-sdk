package iavlx

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidwall/btree"
)

func (ts *TreeStore) load() error {
	// collect a list of all existing subdirectories
	dirs, err := os.ReadDir(ts.dir)
	if err != nil {
		return fmt.Errorf("failed to read tree store dir: %w", err)
	}

	// directory map: startVersion -> compactedAt -> dirName
	var dirMap btree.Map[uint64, *btree.Map[uint64, string]]
	for _, de := range dirs {
		if !de.IsDir() {
			continue
		}
		dir := filepath.Join(ts.dir, de.Name())
		startVersion, compactedAt, valid := ParseChangesetDirName(dir)
		if !valid {
			continue
		}
		if _, found := dirMap.Get(startVersion); !found {
			dirMap.Set(startVersion, &btree.Map[uint64, string]{})
		}
		caMap, _ := dirMap.Get(startVersion)
		caMap.Set(compactedAt, dir)
	}

	// load changesets in order
	for {
		startVersion, compactionMap, ok := dirMap.PopMin()
		if !ok {
			return nil
		}

		lastCompaction, dirName, ok := compactionMap.PopMax()
		if !ok {
			return fmt.Errorf("internal error: no changeset entries for start version %d", startVersion)
		}

		ready, err := IsChangesetReady(dirName)
		if err != nil {
			return fmt.Errorf("failed to check if changeset %s is ready: %w", dirName, err)
		}

		if !ready {
			ts.logger.Warn("found incomplete compaction, deleting", "dir", dirName)
			err := os.RemoveAll(dirName)
			if err != nil {
				ts.logger.Error("failed to remove incomplete compaction", "dir", dirName, "error", err)
			}
			continue
		}

		cf, err := ReopenChangesetFiles(dirName)
	}
}
