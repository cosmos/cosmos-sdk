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

func RunCompactor(ctx context.Context, treeStore *TreeStore, options PruneOptions) error {
	cp := newCompactorProc(treeStore, options)
	return cp.startCompactionRun(ctx)
}

func newCompactorProc(treeStore *TreeStore, options PruneOptions) *CompactorProc {
	return &CompactorProc{
		treeStore: treeStore,
		options:   options,
	}
}

func (cp *CompactorProc) startCompactionRun(ctx context.Context) error {
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

	// then calculate the first compaction to be retained
	info, err := cp.treeStore.checkpointForVersion(cp.options.RetainVersion)
	if err != nil {
		return fmt.Errorf("failed to determine checkpoint for version %d: %w", cp.options.RetainVersion, err)
	}

	if info.Checkpoint == 0 {
		// no checkpoint found for the oldest retained version, nothing to compact
		return nil
	}
	retainVersion := info.Version

	cpOpts := CompactOptions{
		RetainCriteria: func(createCheckpoint, orphanVersion uint32) bool {
			// retain if the node was orphaned at or after the version of the oldest retained checkpoint
			return orphanVersion >= retainVersion
		},
		CompactedAt: cp.treeStore.LatestVersion(),
		// we start the WAL after the oldest retained checkpoint version
		WALStartVersion: retainVersion + 1,
	}

	for _, cs := range toProcess {
		if ctx.Err() != nil {
			// context cancelled, stop processing further changesets
			break
		}
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
