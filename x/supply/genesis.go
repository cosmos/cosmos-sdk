package supply

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets supply information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetSupply(ctx, data.Supply)
	keeper.SetSendEnabled(ctx, data.SendEnabled)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(keeper.GetSupply(ctx), keeper.GetSendEnabled(ctx))
}

// ValidateGenesis performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if err := data.Supply.ValidateBasic(); err != nil {
		return err
	}
	return nil
}
