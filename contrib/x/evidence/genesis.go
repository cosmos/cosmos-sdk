package evidence

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/contrib/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/contrib/x/evidence/keeper"
	types2 "github.com/cosmos/cosmos-sdk/contrib/x/evidence/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the evidence module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs *types2.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types2.ModuleName, err))
	}

	for _, e := range gs.Evidence {
		evi, ok := e.GetCachedValue().(exported.Evidence)
		if !ok {
			panic("expected evidence")
		}
		if _, err := k.Evidences.Get(ctx, evi.Hash()); err == nil {
			panic(fmt.Sprintf("evidence with hash %s already exists", evi.Hash()))
		}

		if err := k.Evidences.Set(ctx, evi.Hash(), evi); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the evidence module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types2.GenesisState {
	gs := new(types2.GenesisState)
	err := k.Evidences.Walk(ctx, nil, func(_ []byte, value exported.Evidence) (stop bool, err error) {
		anyEvi, err := codectypes.NewAnyWithValue(value)
		if err != nil {
			return false, err
		}
		gs.Evidence = append(gs.Evidence, anyEvi)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return gs
}
