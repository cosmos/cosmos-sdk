package types

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params Params, validators []Validator, delegations []Delegation) *GenesisState {
	return &GenesisState{
		Params:      params,
		Validators:  validators,
		Delegations: delegations,
	}
}
