package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
		startVersion, compactedAt, valid := ParseChangesetDirName(dirName)
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

		// startVersion should be equal to stagedVersion

		stagedVersion := ts.StagedVersion()
		if startVersion < stagedVersion {
			logger.WarnContext(ctx, "found undeleted changeset that was already compacted", "startVersion", startVersion, "stagedVersion", stagedVersion)
			// TODO delete undeleted compactions
			continue
		}

		if startVersion > stagedVersion {
			return fmt.Errorf("missing changeset for staged version %d", stagedVersion)
		}

		for {
			_, dirName, ok := compactionMap.PopMax()
			if !ok {
				return fmt.Errorf("internal error: no changeset entries for start version %d", startVersion)
			}

			ready, err := IsChangesetReady(dirName)
			if err != nil {
				return fmt.Errorf("failed to check if changeset %s is ready: %w", dirName, err)
			}

			if !ready {
				logger.WarnContext(ctx, "found incomplete compaction, deleting", "dir", dirName)
				err := os.RemoveAll(dirName)
				if err != nil {
					logger.Error("failed to remove incomplete compaction", "dir", dirName, "error", err)
				}
				continue
			}

			logger.DebugContext(ctx, "loading changeset", "startVersion", startVersion, "dir", dirName)

			cs, err := OpenChangeset(ts, dirName)
			if err != nil {
				return fmt.Errorf("failed to open changeset in %s: %w", dirName, err)
			}

			// TODO fault tolerance checks
			info := cs.files.Info()
			//realStartVersion := cs.info.StartVersion
			//if uint64(realStartVersion) != startVersion {
			//	if realStartVersion == 0 {
			//		if dirMap.Len() != 0 {
			//			return fmt.Errorf("found incomplete changeset %s, but there are later changesets present", dirName)
			//		}
			//		ts.logger.Debug("found final incomplete changeset, deleting", "dir", dirName)
			//		err := os.RemoveAll(dirName)
			//		if err != nil {
			//			return fmt.Errorf("failed to remove incomplete changeset %s: %w", dirName, err)
			//		}
			//		break
			//	} else {
			//		return fmt.Errorf("changeset in %s has mismatched start version %d (expected %d)", dirName, realStartVersion, startVersion)
			//	}
			//}

			firstCheckpoint := info.FirstCheckpoint
			if firstCheckpoint != 0 {
				ts.checkpointer.changesetsByCheckpoint.Set(firstCheckpoint, cs)
			}

			ts.version.Store(info.WALEndVersion)

			break
		}
	}

	root, err := ts.checkpointer.LoadRoot(ts.version.Load())
	if err != nil {
		return fmt.Errorf("failed to load root after loading changesets: %w", err)
	}
	ts.root.Store(root)
	return nil
}
