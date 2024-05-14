package types

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter Minter, params Params, reductionStartedEpoch int64) *GenesisState {
	return &GenesisState{
		Minter:                minter,
		Params:                params,
		ReductionStartedEpoch: reductionStartedEpoch,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Minter:                DefaultInitialMinter(),
		Params:                DefaultParams(),
		ReductionStartedEpoch: 0,
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return ValidateMinter(data.Minter)
}
