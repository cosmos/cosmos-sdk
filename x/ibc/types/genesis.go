package types

import (
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
