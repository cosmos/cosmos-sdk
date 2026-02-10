package internal

import (
	"context"
	"fmt"
)

type compactorProc struct {
	treeStore    *TreeStore
	curCompactor *Compactor
	options      PruneOptions
}

func newCompactorProc(treeStore *TreeStore, options PruneOptions) *compactorProc {
	return &compactorProc{
		treeStore: treeStore,
		options:   options,
	}
}

func (cp *compactorProc) StartCompactionRun(ctx context.Context) error {
	// collect current entries
	toProcess := make([]*Changeset, 0, cp.treeStore.changesetsByVersion.Len())
	cp.treeStore.changesetsLock.RLock()
	cp.treeStore.changesetsByVersion.Ascend(0, func(_ uint32, cs *Changeset) bool {
		toProcess = append(toProcess, cs)
		return true
	})
	cp.treeStore.changesetsLock.RUnlock()

	// first calculate the oldest version to be retained
	latestVersion := cp.treeStore.LatestVersion()
	if cp.options.KeepRecent >= latestVersion {
		// nothing to compact
		return nil
	}
	oldestRetainedVersion := latestVersion - cp.options.KeepRecent

	// then calculate the first compaction to be retained
	info, err := cp.treeStore.checkpointForVersion(oldestRetainedVersion)
	if err != nil {
		return fmt.Errorf("failed to determine checkpoint for version %d: %w", oldestRetainedVersion, err)
	}

	oldestRetainedCheckpoint := uint32(0)
	walRetainVersion := uint32(0)
	if info != nil {
		oldestRetainedCheckpoint = info.Checkpoint
		walRetainVersion = info.Version
	}

	cpOpts := CompactOptions{
		RetainCriteria: func(createCheckpoint, orphanVersion uint32) bool {
			// retain if the node was orphaned at or after the oldest retained checkpoint
			return orphanVersion >= oldestRetainedCheckpoint
		},
		CompactedAt:     latestVersion,
		WALStartVersion: walRetainVersion,
	}

	for _, cs := range toProcess {
		err := cp.compactOne(ctx, cs, cpOpts)
		if err != nil {
			return fmt.Errorf("failed to compact changeset starting at version %d: %w", cs.Files().StartVersion(), err)
		}
	}

	return nil
}

func (cp *compactorProc) compactOne(ctx context.Context, cs *Changeset, opts CompactOptions) error {
	rdr, pin := cs.TryPinReader()
	defer pin.Unpin()
	if rdr == nil {
		return fmt.Errorf("changeset reader is not available for version %d", cs.Files().StartVersion())
	}

	if cp.curCompactor == nil {
		var err error
		cp.curCompactor, err = NewCompactor(ctx, rdr, opts, cp.treeStore)
		if err != nil {
			return fmt.Errorf("failed to create compactor for version %d: %w", cs.Files().StartVersion(), err)
		}
	} else {
		err := cp.curCompactor.AddChangeset(rdr)
		if err != nil {
			return fmt.Errorf("failed to add changeset starting at version %d to compactor: %w", cs.Files().StartVersion(), err)
		}

		// TODO check if we want to roll over the compactor after a certain size
	}
	return nil
}
