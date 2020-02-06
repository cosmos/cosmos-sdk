package types

// PruningOptions defines the specific pruning strategy every store in a multi-store will use
// when committing state, where keepEvery determines which committed heights are flushed
// to disk and snapshotEvery determines which of these heights are kept after pruning.
// Note, that the invariant keepEvery % snapshotEvery = 0 must hold
type PruningOptions struct {
	KeepEvery     int64
	SnapshotEvery int64
}

// Checks if PruningOptions is valid
// SnapshotEvery is a multiple of KeepEvery
func (po PruningOptions) IsValid() bool {
	// Must flush at positive block interval
	if po.KeepEvery <= 0 {
		return false
	}
	// SnapshotEvery must be a multiple of KeepEvery
	return po.SnapshotEvery%po.KeepEvery == 0
}

func (po PruningOptions) FlushVersion(ver int64) bool {
	return po.KeepEvery != 0 && ver%po.KeepEvery == 0
}

func (po PruningOptions) SnapshotVersion(ver int64) bool {
	return po.SnapshotEvery != 0 && ver%po.SnapshotEvery == 0
}

// default pruning strategies
var (
	// PruneEverything means all saved states will be deleted, storing only the current state
	PruneEverything = PruningOptions{
		KeepEvery:     1,
		SnapshotEvery: 0,
	}
	// PruneNothing means all historic states will be saved, nothing will be deleted
	PruneNothing = PruningOptions{
		KeepEvery:     1,
		SnapshotEvery: 1,
	}
	// PruneSyncable means only those states not needed for state syncing will be deleted (flush every 100 + Snapshot every 10000th)
	PruneSyncable = PruningOptions{
		KeepEvery:     100,
		SnapshotEvery: 10000,
	}
)
