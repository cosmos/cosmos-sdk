package types

import (
	"fmt"
)

// GenesisState defines the ibc channel submodule's genesis state.
type GenesisState struct {
	Channels []Channel `json:"channels" yaml:"channels"`
}

// NewGenesisState creates a GenesisState instance.
func NewGenesisState(
	channels []Channel,
) GenesisState {
	return GenesisState{
		Channels: channels,
	}
}

// DefaultGenesisState returns the ibc channel submodule's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Channels: []Channel{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for i, channel := range gs.Channels {
		if err := channel.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid channel %d: %w", i, err)
		}
	}

	return nil
}
