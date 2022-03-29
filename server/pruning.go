package server

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/server/types"
	pruningTypes "github.com/cosmos/cosmos-sdk/pruning/types"
)

// GetPruningOptionsFromFlags parses command flags and returns the correct
// PruningOptions. If a pruning strategy is provided, that will be parsed and
// returned, otherwise, it is assumed custom pruning options are provided.
func GetPruningOptionsFromFlags(appOpts types.AppOptions) (*pruningTypes.PruningOptions, error) {
	strategy := strings.ToLower(cast.ToString(appOpts.Get(FlagPruning)))

	switch strategy {
	case pruningTypes.PruningOptionDefault, pruningTypes.PruningOptionNothing, pruningTypes.PruningOptionEverything:
		return pruningTypes.NewPruningOptionsFromString(strategy), nil

	case pruningTypes.PruningOptionCustom:
		opts := pruningTypes.NewCustomPruningOptions(
			cast.ToUint64(appOpts.Get(FlagPruningKeepRecent)),
			cast.ToUint64(appOpts.Get(FlagPruningInterval)),
		)

		if err := opts.Validate(); err != nil {
			return opts, fmt.Errorf("invalid custom pruning options: %w", err)
		}

		return opts, nil

	default:
		return nil, fmt.Errorf("unknown pruning strategy %s", strategy)
	}
}
