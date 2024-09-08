package types

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// DefaultGenesisState returns a default bank/v2 module genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams())
}

func (gs *GenesisState) Validate() error {
	return gs.Params.Validate()
}
