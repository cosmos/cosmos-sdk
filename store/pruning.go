package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// default pruning strategies
var (
	// PruneEverything means all saved states will be deleted, storing only the current state
	PruneEverything = sdk.NewPruningOptions(0, 0)
	// PruneNothing means all historic states will be saved, nothing will be deleted
	PruneNothing = sdk.NewPruningOptions(0, 1)
	// PruneSyncable means only those states not needed for state syncing will be deleted (keeps last 100 + every 10000th)
	PruneSyncable = sdk.NewPruningOptions(100, 10000)
)

func NewPruningOptions(strategy string) (opt PruningOptions) {
	switch strategy {
	case "nothing":
		opt = PruneNothing
	case "everything":
		opt = PruneEverything
	case "syncable":
		opt = PruneSyncable
	default:
		opt = PruneSyncable
	}
	return
}
