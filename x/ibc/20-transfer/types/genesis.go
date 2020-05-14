package types

// GenesisState is currently only used to ensure that the InitGenesis gets run
// by the module manager
type GenesisState struct {
	PortID string `json:"port_id" yaml:"port_id"`
}

// DefaultGenesis returns a GenesisState with "transfer" as the default PortID.
func DefaultGenesis() GenesisState {
	return GenesisState{
		PortID: PortID,
	}
}
