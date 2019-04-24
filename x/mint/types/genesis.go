package types

// GenesisState - minter state
type GenesisState struct {
	Minter Minter `json:"minter"` // minter object
	Params Params `json:"params"` // inflation params
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter Minter, params Params) GenesisState {
	return GenesisState{
		Minter: minter,
		Params: params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Minter: DefaultInitialMinter(),
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	err := validateParams(data.Params)
	if err != nil {
		return err
	}
	err = validateMinter(data.Minter)
	if err != nil {
		return err
	}
	return nil
}
