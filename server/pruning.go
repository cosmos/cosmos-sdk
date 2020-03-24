package server

import (
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/store"
)

// GetPruningOptionsFromFlags parses start command flags and returns the correct PruningOptions.
// flagPruning prevails over flagPruningKeepEvery and flagPruningSnapshotEvery.
// Default option is PruneSyncable.
func GetPruningOptionsFromFlags() store.PruningOptions {
	if viper.IsSet(flagPruning) {
		return store.NewPruningOptionsFromString(viper.GetString(flagPruning))
	}

	if viper.IsSet(flagPruningKeepEvery) && viper.IsSet(flagPruningSnapshotEvery) {
		return store.PruningOptions{
			KeepEvery:     viper.GetInt64(flagPruningKeepEvery),
			SnapshotEvery: viper.GetInt64(flagPruningSnapshotEvery),
		}
	}

	return store.PruneSyncable
}
