package evidence

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the evidence module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, gs GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	for _, e := range gs.Evidence {
		if _, ok := k.GetEvidence(ctx, e.Hash()); ok {
			panic(fmt.Sprintf("evidence with hash %s already exists", e.Hash()))
		}

		k.SetEvidence(ctx, e)
	}

	k.SetParams(ctx, gs.Params)
}

// ExportGenesis returns the evidence module's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{
		Params:   k.GetParams(ctx),
		Evidence: k.GetAllEvidence(ctx),
	}
}
