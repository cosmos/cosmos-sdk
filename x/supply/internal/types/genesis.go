package types

// GenesisState is the supply state that must be provided at genesis.
type GenesisState struct {
	Supply Supply `json:"supply" yaml:"supply"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(supply Supply) GenesisState {
	return GenesisState{supply}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultSupply())
}

// ValidateGenesis performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return data.Supply.ValidateBasic()
}
