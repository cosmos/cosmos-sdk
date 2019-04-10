package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	SendEnabled       bool          `json:"send_enabled"`
	CirculatingSupply sdk.Coins     `json:"circulating_supply"`
	ModulesHoldings   []TokenHolder `json:"modules_holdings"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(
	sendEnabled bool, circulatingSupply sdk.Coins, holdings []TokenHolder,
) GenesisState {
	return GenesisState{
		SendEnabled:       sendEnabled,
		CirculatingSupply: circulatingSupply,
		ModulesHoldings:   holdings,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(true, sdk.Coins{}, []TokenHolder{})
}

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetSendEnabled(ctx, data.SendEnabled)
	keeper.SetCirculatingSupply(ctx, data.CirculatingSupply)
	for _, holder := range data.ModulesHoldings {
		keeper.SetTokenHolder(ctx, holder.Module)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(
		keeper.GetSendEnabled(ctx),
		keeper.GetCirculatingSupply(ctx),
		keeper.GetTokenHolders(ctx),
	)
}

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if !data.CirculatingSupply.IsValid() {
		return fmt.Errorf("circulating supply coins are not valid")
	}
	return nil
}
