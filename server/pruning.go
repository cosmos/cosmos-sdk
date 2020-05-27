package server

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/store"
)

// GetPruningOptionsFromFlags parses start command flags and returns the correct PruningOptions.
// flagPruning prevails over flagPruningKeepEvery and flagPruningSnapshotEvery.
// Default option is PruneSyncable.
func GetPruningOptionsFromFlags() (store.PruningOptions, error) {
	strategy := viper.GetString(flagPruning)
	switch strategy {
	case "syncable", "nothing", "everything":
		return store.NewPruningOptionsFromString(viper.GetString(flagPruning)), nil

	case "custom":
		opts := store.PruningOptions{
			KeepEvery:     viper.GetInt64(flagPruningKeepEvery),
			SnapshotEvery: viper.GetInt64(flagPruningSnapshotEvery),
		}
		if !opts.IsValid() {
			return opts, fmt.Errorf("invalid granular options")
		}
		return opts, nil

	default:
		return store.PruningOptions{}, fmt.Errorf("unknown pruning strategy %s", strategy)
	}
}
