package types

import (
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewGenesisState creates a new ibc-transfer GenesisState instance.
func NewGenesisState(portID string, denomTraces Traces) *GenesisState {
	return &GenesisState{
		PortID:      portID,
		DenomTraces: denomTraces,
	}
}

// DefaultGenesisState returns a GenesisState with "transfer" as the default PortID.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		PortID:      PortID,
		DenomTraces: Traces{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := host.PortIdentifierValidator(gs.PortID); err != nil {
		return err
	}
	return gs.DenomTraces.Validate()
}
