package internal

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CompactorProc orchestrates a single compaction run across all sealed changesets in a tree.
//
// It iterates changesets in version order, feeding each into a Compactor. When the compacted
// output reaches CompactionRolloverSize, it seals the current Compactor and starts a new one.
// This prevents compacted changesets from growing unboundedly.
//
// The retain criteria is derived from the RetainVersion option: we find the checkpoint at or
// before that version, and keep any node orphaned at or after that checkpoint's version.
// Everything older is prunable.
type CompactorProc struct {
	treeStore    *TreeStore
	curCompactor *Compactor
	options      CompactionOptions
}

// RunCompactor runs a single compaction pass over all sealed changesets in the tree store.
// Called from the background compaction goroutine (see compactIfNeeded in commit_multi_tree.go).
//
// TODO: we could optimize this by pre-scanning each changeset's orphan file to estimate how many
// nodes would be pruned, and skip changesets where the prune count is too low to justify the IO.
// The building block already exists: OrphanRewriter.Preprocess (called in Compactor.doAddChangeset)
// reads the orphan file and returns a deleteMap of prunable nodes. To implement this optimization:
//  1. Before calling compactOne, call OrphanRewriter.Preprocess on the changeset's orphan file
//     with the same RetainCriteria to get len(deleteMap) — the exact prune count.
//  2. Compare that against the total node count (leaves + branches across all checkpoints).
//  3. Skip the changeset if the ratio is below some threshold (e.g. <5% prunable).
//  4. Pass the already-computed deleteMap into doAddChangeset to avoid scanning orphans twice.
// The orphan file is fixed-size entries so scanning it is cheap relative to the full compaction.
func RunCompactor(ctx context.Context, treeStore *TreeStore, options CompactionOptions) error {
	cp := newCompactorProc(treeStore, options)
	return cp.startCompactionRun(ctx)
}

func newCompactorProc(treeStore *TreeStore, options CompactionOptions) *CompactorProc {
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

	cpOpts := CompactorOptions{
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
			// context canceled, stop processing further changesets
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

func (cp *CompactorProc) compactOne(ctx context.Context, cs *Changeset, opts CompactorOptions) error {
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

	if int64(cp.curCompactor.TotalBytes()) > cp.options.CompactionRolloverSize {
		_, err := cp.curCompactor.Seal()
		if err != nil {
			return fmt.Errorf("failed to seal compactor after reaching rollover size: %w", err)
		}
		cp.curCompactor = nil

	}

	return nil
}
