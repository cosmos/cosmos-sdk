package types

import (
	"fmt"

	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

// NewConnectionPaths creates a ConnectionPaths instance.
func NewConnectionPaths(id string, paths []string) ConnectionPaths {
	return ConnectionPaths{
		ClientId: id,
		Paths:    paths,
	}
}

// NewGenesisState creates a GenesisState instance.
func NewGenesisState(
	connections []IdentifiedConnection, connPaths []ConnectionPaths,
	nextConnectionSequence uint64,
) GenesisState {
	return GenesisState{
		Connections:            connections,
		ClientConnectionPaths:  connPaths,
		NextConnectionSequence: nextConnectionSequence,
	}
}

// DefaultGenesisState returns the ibc connection submodule's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Connections:            []IdentifiedConnection{},
		ClientConnectionPaths:  []ConnectionPaths{},
		NextConnectionSequence: 0,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for i, conn := range gs.Connections {
		if err := conn.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid connection %v index %d: %w", conn, i, err)
		}
	}

	for i, conPaths := range gs.ClientConnectionPaths {
		if err := host.ClientIdentifierValidator(conPaths.ClientId); err != nil {
			return fmt.Errorf("invalid client connection path %d: %w", i, err)
		}
		for _, connectionID := range conPaths.Paths {
			if err := host.ConnectionIdentifierValidator(connectionID); err != nil {
				return fmt.Errorf("invalid client connection ID (%s) in connection paths %d: %w", connectionID, i, err)
			}
		}
	}

	return nil
}
