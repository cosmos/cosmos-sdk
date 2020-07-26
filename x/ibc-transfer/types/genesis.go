package types

import (
	host "github.com/KiraCore/cosmos-sdk/x/ibc/24-host"
)

// GenesisState is currently only used to ensure that the InitGenesis gets run
// by the module manager
type GenesisState struct {
	PortID string `json:"port_id" yaml:"port_id"`
}

// DefaultGenesisState returns a GenesisState with "transfer" as the default PortID.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		PortID: PortID,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return host.PortIdentifierValidator(gs.PortID)
}
