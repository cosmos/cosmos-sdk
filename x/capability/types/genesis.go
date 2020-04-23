package types

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// GenesisState represents the Capability module genesis state
type GenesisState struct {
	// capability global index
	Index uint64 `json:"index" yaml:"index"`
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() GenesisState {
	return GenesisState{Index: DefaultIndex}
}
