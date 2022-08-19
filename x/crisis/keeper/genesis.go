package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// new crisis genesis
func (k *Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := k.SetConstantFee(ctx, data.ConstantFee); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k *Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	constantFee := k.GetConstantFee(ctx)
	return types.NewGenesisState(constantFee)
}

func (k Keeper) InitGenesisFrom(ctx sdk.Context, cdc codec.JSONCodec, importPath string) error {
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
	k.SetConstantFee(ctx, gs.ConstantFee)
	return nil
}

func (k Keeper) ExportGenesisTo(ctx sdk.Context, cdc codec.JSONCodec, exportPath string) error {
	f, err := module.CreateGenesisExportFile(exportPath, types.ModuleName)
	if err != nil {
		return err
	}
	defer f.Close()

	gs := k.ExportGenesis(ctx)
	bz := cdc.MustMarshalJSON(gs)
	return module.FileWrite(f, bz)
}
