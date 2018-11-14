package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Pool                 Pool                  `json:"pool"`
	Params               Params                `json:"params"`
	IntraTxCounter       int16                 `json:"intra_tx_counter"`
	LastTotalPower       sdk.Int               `json:"last_total_power"`
	Validators           []Validator           `json:"validators"`
	Bonds                []Delegation          `json:"bonds"`
	UnbondingDelegations []UnbondingDelegation `json:"unbonding_delegations"`
	Redelegations        []Redelegation        `json:"redelegations"`
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
