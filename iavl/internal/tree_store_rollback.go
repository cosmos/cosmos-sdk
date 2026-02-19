package internal

import "fmt"

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
