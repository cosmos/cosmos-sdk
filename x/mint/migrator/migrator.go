package migrator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Migrator is a struct for handling in-place state migrations.
type Migrator struct {
	keeper         keeper.Keeper
	legacySubspace paramstypes.Subspace
}

func New(k keeper.Keeper, ss paramstypes.Subspace) Migrator {
	return Migrator{
		keeper:         k,
		legacySubspace: ss,
	}
}

// Migrate1to2 migrates the x/mint module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/mint
// module state.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	var currParams types.Params
	m.legacySubspace.GetParamSet(ctx, &currParams)

	return m.keeper.SetParams(ctx, currParams)
}
