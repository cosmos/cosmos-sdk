package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
)

// GenesisState defines the ibc module's genesis state.
type GenesisState struct {
	ConnectionGenesis connection.GenesisState `json:"connection_genesis" yaml:"connection_genesis"`
}

// DefaultGenesisState returns the ibc module's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		ConnectionGenesis: connection.DefaultGenesisState(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return gs.ConnectionGenesis.Validate()
}

// InitGenesis initializes the ibc connection submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, gs GenesisState) {
	connection.InitGenesis(ctx, k.ConnectionKeeper, gs.ConnectionGenesis)
}

// ExportGenesis returns the ibc connection submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{
		ConnectionGenesis: connection.ExportGenesis(ctx, k.ConnectionKeeper),
	}
}
