package internal

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type CompactorProc struct {
	treeStore    *TreeStore
	curCompactor *Compactor
	options      PruneOptions
}

func NewCompactorProc(treeStore *TreeStore, options PruneOptions) *CompactorProc {
	return &CompactorProc{
		treeStore: treeStore,
		options:   options,
	}
}

func (cp *CompactorProc) StartCompactionRun(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "CompactorProc.StartCompactionRun")
	defer span.End()

	// collect current entries
	toProcess := make([]*Changeset, 0, cp.treeStore.changesetsByVersion.Len())
	cp.treeStore.changesetsLock.RLock()
	cp.treeStore.changesetsByVersion.Ascend(0, func(_ uint32, cs *Changeset) bool {
		if cs.sealed.Load() {
			toProcess = append(toProcess, cs)
		}
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

	if info.Checkpoint == 0 {
		// no checkpoint found for the oldest retained version, nothing to compact
		return nil
	}
	oldestRetainedCheckpoint := info.Checkpoint
	walRetainVersion := info.Version

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

	if cp.curCompactor != nil {
		_, err := cp.curCompactor.Seal()
		if err != nil {
			return fmt.Errorf("failed to seal compactor after processing all changesets: %w", err)
		}
	}

	return nil
}

func (cp *CompactorProc) compactOne(ctx context.Context, cs *Changeset, opts CompactOptions) error {
	ctx, span := tracer.Start(ctx, "CompactorProc.compactOne",
		trace.WithAttributes(
			attribute.String("changesetDir", cs.Files().Dir()),
		),
	)
	defer span.End()

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
	}

	if cp.curCompactor.TotalBytes() > cp.options.CompactionRolloverSize {
		_, err := cp.curCompactor.Seal()
		if err != nil {
			return fmt.Errorf("failed to seal compactor after reaching rollover size: %w", err)
		}
		cp.curCompactor = nil

	}

	return nil
}
