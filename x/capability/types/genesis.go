package types

// GenesisState represents the Capability module genesis state
type GenesisState struct {
	Index uint64 `json:"index" yaml:"index"`
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() GenesisState {
	return GenesisState{Index: 1}
}
