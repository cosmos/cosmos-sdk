package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	SendEnabled bool `json:"send_enabled" yaml:"send_enabled"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(sendEnabled bool) GenesisState {
	return GenesisState{SendEnabled: sendEnabled}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState { return NewGenesisState(true) }

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetSendEnabled(ctx, data.SendEnabled)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(keeper.GetSendEnabled(ctx))
}

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error { return nil }
