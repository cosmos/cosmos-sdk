package types

// NewGenesisState creates new GenesisState object
func NewGenesisState(entries []MsgGrantAuthorization) *GenesisState {
	return &GenesisState{
		Authorization: entries,
	}
}

// ValidateGenesis check the given genesis state has no integrity issues
func ValidateGenesis(data GenesisState) error {
	return nil
}

//
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}
