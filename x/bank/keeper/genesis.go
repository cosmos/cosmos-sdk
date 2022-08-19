package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// InitGenesis initializes the bank module's state from a given genesis state.
func (k BaseKeeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	for _, se := range genState.GetAllSendEnabled() {
		k.SetSendEnabled(ctx, se.Denom, se.Enabled)
	}

	totalSupply := sdk.Coins{}
	genState.Balances = types.SanitizeGenesisBalances(genState.Balances)

	for _, balance := range genState.Balances {
		addr := balance.GetAddress()

		if err := k.initBalances(ctx, addr, balance.Coins); err != nil {
			panic(fmt.Errorf("error on setting balances %w", err))
		}

		totalSupply = totalSupply.Add(balance.Coins...)
	}

	if !genState.Supply.Empty() && !genState.Supply.IsEqual(totalSupply) {
		panic(fmt.Errorf("genesis supply is incorrect, expected %v, got %v", genState.Supply, totalSupply))
	}

	for _, supply := range totalSupply {
		k.setSupply(ctx, supply)
	}

	for _, meta := range genState.DenomMetadata {
		k.SetDenomMetaData(ctx, meta)
	}
}

// ExportGenesis returns the bank module's genesis state.
func (k BaseKeeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	totalSupply, _, err := k.GetPaginatedTotalSupply(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(fmt.Errorf("unable to fetch total supply %v", err))
	}

	rv := types.NewGenesisState(
		k.GetParams(ctx),
		k.GetAccountsBalances(ctx),
		totalSupply,
		k.GetAllDenomMetaData(ctx),
		k.GetAllSendEnabledEntries(ctx),
	)
	return rv
}

func (k BaseKeeper) ExportGenesisTo(ctx sdk.Context, cdc codec.JSONCodec, exportPath string) error {
	f, err := module.CreateGenesisExportFile(exportPath, types.ModuleName)
	if err != nil {
		return err
	}
	defer f.Close()

	gs := k.ExportGenesis(ctx)
	bz := cdc.MustMarshalJSON(gs)
	return module.FileWrite(f, bz)
}

func (k BaseKeeper) InitGenesisFrom(ctx sdk.Context, cdc codec.JSONCodec, importPath string) error {
	f, err := module.OpenGenesisModuleFile(importPath, types.ModuleName)
	if err != nil {
		return err
	}
	defer f.Close()

	bz, err := module.FileRead(f)
	if err != nil {
		return err
	}

	var gs types.GenesisState
	cdc.MustUnmarshalJSON(bz, &gs)

	k.InitGenesis(ctx, &gs)
	return nil
}
