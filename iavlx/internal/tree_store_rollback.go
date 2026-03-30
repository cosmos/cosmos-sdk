package internal

import "fmt"

// rollbackToVersion rolls a single tree back to a specific version.
//
// NOTE: This method is currently UNUSED. In-process rollback (CommitMultiTree.RollbackToVersion)
// uses the filesystem-level RollbackMultiTree/RollbackTree functions from rollback.go instead,
// then poisons the in-memory state and requires a restart. This method attempted to do an in-memory
// rollback that could continue serving, but that path was never completed.
//
// This code is kept for reference and potential future use.
// It is more destructive than crash recovery — it deletes changeset directories and truncates
// checkpoint files, not just WAL entries. The tree must not be actively committing when this is called.
//
// The rollback proceeds in layers, from coarsest to finest:
//  1. Delete entire changeset directories that start after the target version (they're wholly invalid).
//  2. Roll back checkpoints that are ahead of the target version (truncate checkpoint data files
//     back to the previous checkpoint's offsets, one checkpoint at a time).
//  3. Truncate the WAL in the target version's changeset to remove entries beyond the target version.
//  4. Reconstruct the in-memory root by loading from the latest valid checkpoint + WAL replay.
//  5. Initialize a fresh writer for future commits.
func (ts *TreeStore) rollbackToVersion(version uint32) error {
	// Seal the current writer so no more writes can happen to the current changeset.
	if err := ts.currentWriter.Seal(); err != nil {
		return fmt.Errorf("sealing current writer after rollback: %w", err)
	}

	// Step 1: Delete changeset directories that start after the target version.
	// These contain data exclusively for versions we're rolling back, so they're entirely invalid.
	ts.changesetsLock.Lock()
	var err error
	ts.changesetsByVersion.AscendMut(version, func(csVersion uint32, cs *Changeset) bool {
		if csVersion > version {
			ts.logger.Info("deleting changeset newer than target version during rollback", "changesetVersion", csVersion, "targetVersion", version)
			ts.changesetsByVersion.Delete(csVersion)
			rdr, pin := cs.TryPinUncompactedReader()
			if rdr != nil {
				firstCheckpoint := rdr.FirstCheckpoint()
				if firstCheckpoint != 0 {
					ts.checkpointer.changesetLock.Lock()
					ts.checkpointer.changesetsByCheckpoint.Delete(firstCheckpoint)
					ts.checkpointer.changesetLock.Unlock()
				}
			}
			pin.Unpin()

			err = cs.Close()
			if err != nil {
				ts.logger.Error("failed to close changeset during rollback cleanup", "changesetVersion", csVersion, "error", err)
				return false
			}
			err = cs.Files().DeleteFiles()
			if err != nil {
				ts.logger.Error("failed to delete changeset files during rollback cleanup", "changesetVersion", csVersion, "error", err)
				return false
			}
		}
		return true
	})
	ts.changesetsLock.Unlock()
	if err != nil {
		return fmt.Errorf("failed to delete old changesets during rollback: %w", err)
	}

	// Refresh checkpoint bookkeeping after potentially deleting changesets that contained checkpoints.
	if cpInfo, cpErr := ts.checkpointer.LatestCheckpointRoot(); cpErr == nil {
		ts.lastCheckpointVersion = cpInfo.Version
		ts.lastCheckpoint.Store(cpInfo.Checkpoint)
		ts.checkpointer.savedCheckpoint.Store(cpInfo.Checkpoint)
	} else {
		ts.lastCheckpointVersion = 0
		ts.lastCheckpoint.Store(0)
		ts.checkpointer.savedCheckpoint.Store(0)
	}

	// Step 2: Roll back checkpoints that are ahead of the target version.
	// Each rollbackLastCheckpoint call truncates the checkpoint data files (branches.dat, leaves.dat,
	// kv.dat, checkpoints.dat) back by one checkpoint. We repeat until the latest checkpoint
	// is at or before the target version.
	for ts.lastCheckpointVersion > version {
		_, err := ts.rollbackLastCheckpoint()
		if err != nil {
			return fmt.Errorf("failed to roll back checkpoint during rollback to version: %w", err)
		}
	}

	// Step 3: Find the changeset that contains the target version and truncate its WAL.
	// The WAL may contain entries for versions after the target (written during commits that
	// we're now rolling back). RollbackWAL scans for the last commit entry at or before the
	// target version and truncates everything after it.
	cs := ts.changesetForVersion(version)
	if cs == nil {
		return fmt.Errorf("cannot find changeset for target version %d during rollback", version)
	}

	err = RollbackWAL(cs.Files().WALFile(), uint64(version))
	if err != nil {
		return fmt.Errorf("failed to roll back WAL during rollback to version: %w", err)
	}

	// Step 4: Reconstruct the in-memory root at the target version.
	// This uses the same checkpoint + WAL replay mechanism as normal startup.
	root, err := ts.RootAtVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get root at target version %d during rollback: %w", version, err)
	}

	ts.root.Store(&versionedRoot{
		version: version,
		root:    root,
	})

	// Step 5: Initialize a fresh writer so new commits can proceed.
	if err := ts.initNewWriter(); err != nil {
		return fmt.Errorf("reinitializing writer after rollback: %w", err)
	}

	return nil
}

// rollbackLastCheckpoint rolls back the latest checkpoint by one version.
// It updates the checkpointer state and removes any stale checkpoint map entries if necessary.
// The tree must not be actively committing when is called.
// If the rollback is successful, it returns the new latest checkpoint root info after the rollback,
// which should be used to update the tree's in-memory state to reflect the rollback.
func (ts *TreeStore) rollbackLastCheckpoint() (CheckpointRootInfo, error) {
	// get last checkpoint
	latestCheckpoint := ts.checkpointer.savedCheckpoint.Load()
	if latestCheckpoint == 0 {
		return CheckpointRootInfo{}, fmt.Errorf("no checkpoints to roll back")
	}
	// get the changeset for this checkpoint
	cs := ts.checkpointer.ChangesetByCheckpoint(latestCheckpoint)
	if cs == nil {
		return CheckpointRootInfo{}, fmt.Errorf("cannot find changeset for latest checkpoint %d to roll back", latestCheckpoint)
	}
	// pin the changeset's reader
	rdr, pin := cs.TryPinUncompactedReader()
	defer pin.Unpin()
	if rdr == nil {
		return CheckpointRootInfo{}, fmt.Errorf("changeset reader not available for checkpoint rollback")
	}
	// rollback checkpoint
	err := cs.RollbackLastCheckpoint(rdr)
	if err != nil {
		return CheckpointRootInfo{}, fmt.Errorf("failed to roll back latest checkpoint: %w", err)
	}

	// update checkpointer state after rollback
	ts.checkpointer.savedCheckpoint.Store(latestCheckpoint - 1)
	ts.lastCheckpoint.Store(latestCheckpoint - 1)
	// delete the checkpoint from the map if it exists, since it's no longer valid after the rollback
	ts.checkpointer.changesetLock.Lock()
	ts.checkpointer.changesetsByCheckpoint.Delete(latestCheckpoint)
	ts.checkpointer.changesetLock.Unlock()

	// re-get the latest checkpoint root after rollback
	cpInfo, err := ts.checkpointer.LatestCheckpointRoot()
	if err != nil {
		return CheckpointRootInfo{}, fmt.Errorf("failed to get checkpoint root after rollback: %w", err)
	}
	// update last checkpoint version to the version of the new latest checkpoint after rollback
	ts.lastCheckpointVersion = cpInfo.Version

	ts.logger.Info("rolled back last checkpoint", "newLatestCheckpoint", latestCheckpoint-1, "newLatestCheckpointVersion", cpInfo.Version)
	return cpInfo, nil
}
