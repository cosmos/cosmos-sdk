package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

// GenesisState defines the ibc module's genesis state.
type GenesisState struct {
	ClientGenesis     client.GenesisState     `json:"client_genesis" yaml:"client_genesis"`
	ConnectionGenesis connection.GenesisState `json:"connection_genesis" yaml:"connection_genesis"`
	ChannelGenesis    channel.GenesisState    `json:"channel_genesis" yaml:"channel_genesis"`
}

// DefaultGenesisState returns the ibc module's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		ClientGenesis:     client.DefaultGenesisState(),
		ConnectionGenesis: connection.DefaultGenesisState(),
		ChannelGenesis:    channel.DefaultGenesisState(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.ClientGenesis.Validate(); err != nil {
		return err
	}

	if err := gs.ConnectionGenesis.Validate(); err != nil {
		return err
	}

	return gs.ChannelGenesis.Validate()
}

// InitGenesis initializes the ibc state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, createLocalhost bool, gs GenesisState) {
	client.InitGenesis(ctx, k.ClientKeeper, gs.ClientGenesis)
	connection.InitGenesis(ctx, k.ConnectionKeeper, gs.ConnectionGenesis)
	channel.InitGenesis(ctx, k.ChannelKeeper, gs.ChannelGenesis)
}

// ExportGenesis returns the ibc exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{
		ClientGenesis:     client.ExportGenesis(ctx, k.ClientKeeper),
		ConnectionGenesis: connection.ExportGenesis(ctx, k.ConnectionKeeper),
		ChannelGenesis:    channel.ExportGenesis(ctx, k.ChannelKeeper),
	}
}
