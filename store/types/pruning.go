package types

// PruningStrategy specifies how old states will be deleted over time where
// keepRecent can be used with keepEvery to create a pruning "strategy".
type PruningOptions struct {
	keepRecent int64
	keepEvery  int64
}

func NewPruningOptions(keepRecent, keepEvery int64) PruningOptions {
	return PruningOptions{
		keepRecent: keepRecent,
		keepEvery:  keepEvery,
	}
}

// How much recent state will be kept. Older state will be deleted.
func (po PruningOptions) KeepRecent() int64 {
	return po.keepRecent
}

// Keeps every N stated, deleting others.
func (po PruningOptions) KeepEvery() int64 {
	return po.keepEvery
}

// default pruning strategies
var (
	// PruneEverything means all saved states will be deleted, storing only the current state
	PruneEverything = NewPruningOptions(0, 0)
	// PruneNothing means all historic states will be saved, nothing will be deleted
	PruneNothing = NewPruningOptions(0, 1)
	// PruneSyncable means only those states not needed for state syncing will be deleted (keeps last 100 + every 10000th)
	PruneSyncable = NewPruningOptions(100, 10000)
)
