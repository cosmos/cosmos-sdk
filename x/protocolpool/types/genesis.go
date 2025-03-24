package types

func NewGenesisState(cf []ContinuousFund) *GenesisState {
	return &GenesisState{
		ContinuousFunds: cf,
		Params:          Params{},
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ContinuousFunds: []ContinuousFund{},
		Params:          DefaultParams(),
	}
}

// ValidateGenesis validates the genesis state of protocolpool genesis input
func ValidateGenesis(gs *GenesisState) error {
	for _, cf := range gs.ContinuousFunds {
		if err := cf.Validate(); err != nil {
			return err
		}
	}
	return gs.Params.Validate()
}
