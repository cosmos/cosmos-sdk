package server

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"

	"github.com/KiraCore/cosmos-sdk/store"
	"github.com/KiraCore/cosmos-sdk/store/types"
)

// GetPruningOptionsFromFlags parses command flags and returns the correct
// PruningOptions. If a pruning strategy is provided, that will be parsed and
// returned, otherwise, it is assumed custom pruning options are provided.
func GetPruningOptionsFromFlags(appOpts AppOptions) (types.PruningOptions, error) {
	strategy := strings.ToLower(cast.ToString(appOpts.Get(FlagPruning)))

	switch strategy {
	case types.PruningOptionDefault, types.PruningOptionNothing, types.PruningOptionEverything:
		return types.NewPruningOptionsFromString(strategy), nil

	case types.PruningOptionCustom:
		opts := types.NewPruningOptions(
			cast.ToUint64(appOpts.Get(FlagPruningKeepRecent)),
			cast.ToUint64(appOpts.Get(FlagPruningKeepEvery)),
			cast.ToUint64(appOpts.Get(FlagPruningInterval)),
		)

		if err := opts.Validate(); err != nil {
			return opts, fmt.Errorf("invalid custom pruning options: %w", err)
		}

		return opts, nil

	default:
		return store.PruningOptions{}, fmt.Errorf("unknown pruning strategy %s", strategy)
	}
}
