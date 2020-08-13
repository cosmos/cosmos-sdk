package types

import (
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// DefaultGenesisState returns a GenesisState with "transfer" as the default PortID.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		PortID: PortID,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return host.PortIdentifierValidator(gs.PortID)
}
