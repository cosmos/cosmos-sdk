package types

// PruningStrategy specifies how old states will be deleted over time where
// keepRecent can be used with keepEvery to create a pruning "strategy".
// CONTRACT: snapshotEvery is a multiple of keepEvery
type PruningOptions struct {
	keepRecent    int64
	keepEvery     int64
	snapshotEvery int64
}

func NewPruningOptions(keepRecent, keepEvery, snapshotEvery int64) PruningOptions {
	return PruningOptions{
		keepRecent:    keepRecent,
		keepEvery:     keepEvery,
		snapshotEvery: snapshotEvery,
	}
}

// Checks if PruningOptions is valid
// KeepRecent should be larger than KeepEvery
// SnapshotEvery is a multiple of KeepEvery
func (po PruningOptions) IsValid() bool {
	// If we're flushing periodically, we must make in-memory cache less than
	// that period to avoid inefficiency and undefined behavior.
	if po.keepEvery != 0 && po.keepRecent > po.keepEvery {
		return false
	}
	// snapshotEvery must be a multiple of keepEvery
	if po.keepEvery == 0 {
		return po.snapshotEvery == 0
	}
	return po.snapshotEvery%po.keepEvery == 0
}

func (po PruningOptions) FlushVersion(ver int64) bool {
	return po.keepEvery != 0 && ver%po.keepEvery == 0
}

func (po PruningOptions) SnapshotVersion(ver int64) bool {
	return po.snapshotEvery != 0 && ver%po.snapshotEvery == 0
}

// How much recent state will be kept. Older state will be deleted.
func (po PruningOptions) KeepRecent() int64 {
	return po.keepRecent
}

// Keeps every N stated, deleting others.
func (po PruningOptions) KeepEvery() int64 {
	return po.keepEvery
}

func (po PruningOptions) SnapshotEvery() int64 {
	return po.snapshotEvery
}

// default pruning strategies
var (
	// PruneEverything means all saved states will be deleted, storing only the current state
	PruneEverything = NewPruningOptions(0, 1, 0)
	// PruneNothing means all historic states will be saved, nothing will be deleted
	PruneNothing = NewPruningOptions(0, 1, 1)
	// PruneSyncable means only those states not needed for state syncing will be deleted (flush every 100 + snapshot every 10000th)
	PruneSyncable = NewPruningOptions(1, 100, 10000)
)
