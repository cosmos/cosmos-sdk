package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// InitGenesis initializes the bank module's state from a given genesis state.
func (k BaseKeeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	totalSupply := sdk.Coins{}

	genState.Balances = types.SanitizeGenesisBalances(genState.Balances)
	for _, balance := range genState.Balances {
		addr, err := sdk.AccAddressFromBech32(balance.Address)
		if err != nil {
			panic(err)
		}

		if err := k.SetBalances(ctx, addr, balance.Coins); err != nil {
			panic(fmt.Errorf("error on setting balances %w", err))
		}

		totalSupply = totalSupply.Add(balance.Coins...)
	}

	if !genState.Supply.Empty() && !genState.Supply.IsEqual(totalSupply) {
		panic(fmt.Errorf("genesis supply is incorrect, expected %v, got %v", genState.Supply, totalSupply))
	}

<<<<<<< HEAD
	k.SetSupply(ctx, types.NewSupply(genState.Supply))
=======
	k.setSupply(ctx, totalSupply)
>>>>>>> e9e978d54... Fix genesis supply handling (#8930)

	for _, meta := range genState.DenomMetadata {
		k.SetDenomMetaData(ctx, meta)
	}
}

// ExportGenesis returns the bank module's genesis state.
func (k BaseKeeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetAccountsBalances(ctx),
		k.GetSupply(ctx).GetTotal(),
		k.GetAllDenomMetaData(ctx),
	)
}
