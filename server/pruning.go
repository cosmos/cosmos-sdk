package server

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// GetPruningOptionsFromFlags parses command flags and returns the correct
// PruningOptions. If a pruning strategy is provided, that will be parsed and
// returned, otherwise, it is assumed custom pruning options are provided.
func GetPruningOptionsFromFlags(v *viper.Viper) (types.PruningOptions, error) {
	strategy := strings.ToLower(v.GetString(FlagPruning))

	switch strategy {
	case types.PruningOptionDefault, types.PruningOptionNothing, types.PruningOptionEverything:
		return types.NewPruningOptionsFromString(strategy), nil

	case types.PruningOptionCustom:
		opts := types.NewPruningOptions(
			v.GetUint64(FlagPruningKeepRecent),
			v.GetUint64(FlagPruningKeepEvery), v.GetUint64(FlagPruningInterval),
		)

		if err := opts.Validate(); err != nil {
			return opts, fmt.Errorf("invalid custom pruning options: %w", err)
		}

		return opts, nil

	default:
		return store.PruningOptions{}, fmt.Errorf("unknown pruning strategy %s", strategy)
	}
}
