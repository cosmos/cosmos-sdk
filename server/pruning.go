package server

import (
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/spf13/viper"
)

//GetPruningOptionsFromFlags parses start command flags and returns the correct PruningOptions.
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
