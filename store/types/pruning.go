package types

var (
	// PruneEverything defines a pruning strategy where all committed states will
	// be deleted, persisting only the current state.
	PruneEverything = PruningOptions{
		KeepEvery:     1,
		SnapshotEvery: 0,
	}

	// PruneNothing defines a pruning strategy where all committed states will be
	// kept on disk, i.e. no states will be pruned.
	PruneNothing = PruningOptions{
		KeepEvery:     1,
		SnapshotEvery: 1,
	}

	// PruneSyncable defines a pruning strategy where only those states not needed
	// for state syncing will be pruned. It flushes every 100th state to disk and
	// keeps every 10000th.
	PruneSyncable = PruningOptions{
		KeepEvery:     100,
		SnapshotEvery: 10000,
	}
)

// PruningOptions defines the specific pruning strategy every store in a multi-store
// will use when committing state, where keepEvery determines which committed
// heights are flushed to disk and snapshotEvery determines which of these heights
// are kept after pruning.
type PruningOptions struct {
	KeepEvery     int64
	SnapshotEvery int64
}

// IsValid verifies if the pruning options are valid. It returns false if invalid
// and true otherwise. Pruning options are considered valid iff:
//
// - KeepEvery > 0
// - SnapshotEvery >= 0
// - SnapshotEvery % KeepEvery = 0
func (po PruningOptions) IsValid() bool {
	// must flush at positive block interval
	if po.KeepEvery <= 0 {
		return false
	}

	// cannot snapshot negative intervals
	if po.SnapshotEvery < 0 {
		return false
	}

	return po.SnapshotEvery%po.KeepEvery == 0
}

// FlushVersion returns a boolean signaling if the provided version/height should
// be flushed to disk.
func (po PruningOptions) FlushVersion(ver int64) bool {
	return po.KeepEvery != 0 && ver%po.KeepEvery == 0
}

// SnapshotVersion returns a boolean signaling if the provided version/height
// should be snapshotted (kept on disk).
func (po PruningOptions) SnapshotVersion(ver int64) bool {
	return po.SnapshotEvery != 0 && ver%po.SnapshotEvery == 0
}
