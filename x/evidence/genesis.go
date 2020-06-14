package evidence

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

// InitGenesis initializes the evidence module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	for _, e := range gs.Evidence {
		if _, ok := k.GetEvidence(ctx, e.Hash()); ok {
			panic(fmt.Sprintf("evidence with hash %s already exists", e.Hash()))
		}

		k.SetEvidence(ctx, e)
	}
}

// ExportGenesis returns the evidence module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		Evidence: k.GetAllEvidence(ctx),
	}
}
