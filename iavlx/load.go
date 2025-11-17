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

	ts.savedVersion.Store(0)
	ts.stagedVersion = 1
	// load changesets in order
	for {
		startVersion, compactionMap, ok := dirMap.PopMin()
		if !ok {
			return nil
		}

		_, dirName, ok := compactionMap.PopMax()
		if !ok {
			return fmt.Errorf("internal error: no changeset entries for start version %d", startVersion)
		}

		if startVersion < uint64(ts.stagedVersion) {
			ts.logger.Warn("found undeleted compaction", "startVersion", startVersion, "stagedVersion", ts.stagedVersion, "dir", dirName)
			// TODO delete undeleted compactions
			continue
		}

		if startVersion > uint64(ts.stagedVersion) {
			return fmt.Errorf("missing changeset for staged version %d", ts.stagedVersion)
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

		cs, err := OpenChangeset(ts, dirName)
		if err != nil {
			return fmt.Errorf("failed to open changeset in %s: %w", dirName, err)
		}

		ce := &changesetEntry{}
		ce.changeset.Store(cs)
		ts.changesets.Set(uint32(startVersion), ce)

		ts.savedVersion.Store(cs.info.EndVersion)
		ts.stagedVersion = cs.info.EndVersion + 1
	}
}
