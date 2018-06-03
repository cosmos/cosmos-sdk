package types

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Pool       Pool         `json:"pool"`
	Params     Params       `json:"params"`
	Validators []Validator  `json:"validators"`
	Bonds      []Delegation `json:"bonds"`
}

func NewGenesisState(pool Pool, params Params, validators []Validator, bonds []Delegation) GenesisState {
	return GenesisState{
		Pool:       pool,
		Params:     params,
		Validators: validators,
		Bonds:      bonds,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Pool:   InitialPool(),
		Params: DefaultParams(),
	}
}
