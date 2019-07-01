package types

// GenesisState is the supply state that must be provided at genesis.
type GenesisState struct {
	Supply Supply `json:"supply"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(supply Supply) GenesisState {
	return GenesisState{supply}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultSupply())
}
