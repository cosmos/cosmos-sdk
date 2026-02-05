package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/btree"
)

func (ts *TreeStore) load() error {
	ctx, span := tracer.Start(context.Background(), "TreeStore.load")
	defer span.End()

	// collect a list of all existing subdirectories
	dirs, err := os.ReadDir(ts.dir)
	if err != nil {
		return fmt.Errorf("failed to read tree store dir: %w", err)
	}

	// directory map: startVersion -> compactedAt -> dirName
	var dirMap btree.Map[uint32, *btree.Map[uint32, string]]
	for _, de := range dirs {
		if !de.IsDir() {
			continue
		}

		dirName := de.Name()

		// delete -tmp dirs
		if strings.HasSuffix(dirName, "-tmp") {
			dir := filepath.Join(ts.dir, dirName)
			logger.WarnContext(ctx, "found incomplete changeset dir, deleting", "dir", dir)
			err := os.RemoveAll(dir)
			if err != nil {
				logger.ErrorContext(ctx, "failed to remove incomplete changeset dir", "dir", dir, "error", err)
			}
			continue
		}

		startVersion, _, compactedAt, valid := ParseChangesetDirName(dirName)
		if !valid {
			continue
		}

		dir := filepath.Join(ts.dir, dirName)
		if _, found := dirMap.Get(startVersion); !found {
			dirMap.Set(startVersion, &btree.Map[uint32, string]{})
		}
		caMap, _ := dirMap.Get(startVersion)
		caMap.Set(compactedAt, dir)
	}

	// load changesets in order
	for {
		startVersion, compactionMap, ok := dirMap.PopMin()
		if !ok {
			// done
			break
		}
		_, dirName, ok := compactionMap.PopMax()
		if !ok {
			return fmt.Errorf("internal error: no changeset entries for start version %d", startVersion)
		}

		logger.DebugContext(ctx, "loading changeset", "startVersion", startVersion, "dir", dirName)

		cs, err := OpenChangeset(ts, dirName)
		if err != nil {
			return fmt.Errorf("failed to open changeset in %s: %w", dirName, err)
		}

		// store changeset in the tree store by version map
		ts.changesetsByVersion.Set(startVersion, cs)

		// store changeset in the checkpointer by checkpoint map
		rdr, pin := cs.TryPinReader()
		if rdr == nil {
			return fmt.Errorf("changeset reader is not available for changeset starting at version %d", startVersion)
		}
		firstCheckpoint := rdr.FirstCheckpoint()
		if firstCheckpoint != 0 {
			if firstCheckpoint <= ts.checkpointer.savedCheckpoint.Load() {
				return fmt.Errorf("found duplicate or out-of-order checkpoint %d in changeset starting at version %d", firstCheckpoint, startVersion)
			}
			ts.checkpointer.changesetsByCheckpoint.Set(firstCheckpoint, cs)
			lastCheckpoint := rdr.LastCheckpoint()
			// save last checkpoint
			ts.checkpointer.savedCheckpoint.Store(lastCheckpoint)
			ts.checkpoint.Store(lastCheckpoint)
			// save last checkpoint version
			info, err := rdr.GetCheckpointInfo(lastCheckpoint)
			if err != nil {
				return fmt.Errorf("failed to get checkpoint info for checkpoint %d in changeset starting at version %d: %w", lastCheckpoint, startVersion, err)
			}
			ts.lastCheckpointVersion = info.Version
		}
		pin.Unpin()

		// TODO remove older undeleted changesets after compaction
	}

	// get the changeset with the last checkpoint
	root, version, err := ts.checkpointer.LatestCheckpointRoot()
	if err != nil {
		return fmt.Errorf("failed to load root after loading changesets: %w", err)
	}

	// find the changeset to start replaying from
	replayFrom := ts.changesetForVersion(version + 1)
	var replayFromVersion uint32 // default to 0
	if replayFrom != nil {
		replayFromVersion = replayFrom.Files().StartVersion()
	}

	ts.changesetsByVersion.Ascend(replayFromVersion, func(_ uint32, cs *Changeset) bool {
		root, version, err = ReplayWAL(ctx, root, cs.files.WALFile(), version, 0)
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return fmt.Errorf("failed to replay WAL after loading changesets: %w", err)
	}

	ts.root.Store(root)
	ts.version.Store(version)

	return nil
}
