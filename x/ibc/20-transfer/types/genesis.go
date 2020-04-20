package types

// GenesisState is currently only used to ensure that the InitGenesis gets run
// by the module manager
type GenesisState struct {
	PortID string `json:"portid" yaml:"portid"`
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		PortID: PortID,
	}
}
