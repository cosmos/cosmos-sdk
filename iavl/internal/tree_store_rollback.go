package internal

import "fmt"

func (ts *TreeStore) rollbackToVersion(version uint32) error {
	// seal the current writer
	if err := ts.currentWriter.Seal(); err != nil {
		return fmt.Errorf("sealing current writer after rollback: %w", err)
	}

	// delete all changeset that are actually newer than the target version, since they will be invalid after the rollback
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
					// also clear it from the checkpointer map
					ts.checkpointer.changesetLock.Lock()
					ts.checkpointer.changesetsByCheckpoint.Delete(firstCheckpoint)
					ts.checkpointer.changesetLock.Unlock()
				}
			}
			pin.Unpin()

			err = cs.Close() // close the changeset to release any resources
			if err != nil {
				ts.logger.Error("failed to close changeset during rollback cleanup", "changesetVersion", csVersion, "error", err)
				return false // stop iteration on error
			}
			err = cs.Files().DeleteFiles()
			if err != nil {
				ts.logger.Error("failed to delete changeset files during rollback cleanup", "changesetVersion", csVersion, "error", err)
				return false // stop iteration on error
			}
		}
		return true
	})
	ts.changesetsLock.Unlock()
	if err != nil {
		return fmt.Errorf("failed to delete old changesets during rollback: %w", err)
	}

	// refresh last checkpoint version
	if cpInfo, cpErr := ts.checkpointer.LatestCheckpointRoot(); cpErr == nil {
		ts.lastCheckpointVersion = cpInfo.Version
		ts.lastCheckpoint.Store(cpInfo.Checkpoint)
		ts.checkpointer.savedCheckpoint.Store(cpInfo.Checkpoint)
	} else {
		ts.lastCheckpointVersion = 0
		ts.lastCheckpoint.Store(0)
		ts.checkpointer.savedCheckpoint.Store(0)
	}

	// rollback any checkpoints after the target version, since they will be invalid after the rollback
	for ts.lastCheckpointVersion > version {
		_, err := ts.rollbackLastCheckpoint()
		if err != nil {
			return fmt.Errorf("failed to roll back checkpoint during rollback to version: %w", err)
		}
	}

	// now find the changeset which is the latest one that is <= the target version, and roll it back to the target version if necessary
	cs := ts.changesetForVersion(version)
	if cs == nil {
		return fmt.Errorf("cannot find changeset for target version %d during rollback", version)
	}

	// rollback the WAL
	err = RollbackWAL(cs.Files().WALFile(), uint64(version))
	if err != nil {
		return fmt.Errorf("failed to roll back WAL during rollback to version: %w", err)
	}

	root, err := ts.RootAtVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get root at target version %d during rollback: %w", version, err)
	}

	ts.root.Store(&versionedRoot{
		version: version,
		root:    root,
	})

	// initialize a new writer
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
