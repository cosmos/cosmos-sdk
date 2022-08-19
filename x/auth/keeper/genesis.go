package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// InitGenesis - Init store state from genesis data
//
// CONTRACT: old coins from the FeeCollectionKeeper need to be transferred through
// a genesis port script to the new fee collector account
func (ak AccountKeeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := ak.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	accounts, err := types.UnpackAccounts(data.Accounts)
	if err != nil {
		panic(err)
	}
	accounts = types.SanitizeGenesisAccounts(accounts)

	for _, a := range accounts {
		acc := ak.NewAccount(ctx, a)
		ak.SetAccount(ctx, acc)
	}

	ak.GetModuleAccount(ctx, types.FeeCollectorName)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func (ak AccountKeeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := ak.GetParams(ctx)

	var genAccounts types.GenesisAccounts
	ak.IterateAccounts(ctx, func(account types.AccountI) bool {
		genAccount := account.(types.GenesisAccount)
		genAccounts = append(genAccounts, genAccount)
		return false
	})

	return types.NewGenesisState(params, genAccounts)
}

func (ak AccountKeeper) InitGenesisFrom(ctx sdk.Context, cdc codec.JSONCodec, importPath string) error {
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
	ak.InitGenesis(ctx, gs)
	return nil
}

// ExportGenesisTo returns a GenesisState for a given context, keeper and export path
func (ak AccountKeeper) ExportGenesisTo(ctx sdk.Context, cdc codec.JSONCodec, exportPath string) error {
	f, err := module.CreateGenesisExportFile(exportPath, types.ModuleName)
	if err != nil {
		return err
	}
	defer f.Close()

	gs := ak.ExportGenesis(ctx)
	bz := cdc.MustMarshalJSON(gs)
	return module.FileWrite(f, bz)
}
