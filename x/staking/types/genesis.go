package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Pool                 Pool                  `json:"pool"`
	Params               Params                `json:"params"`
	LastTotalPower       sdk.Int               `json:"last_total_power"`
	LastValidatorPowers  []LastValidatorPower  `json:"last_validator_powers"`
	Validators           Validators            `json:"validators"`
	Delegations          Delegations           `json:"delegations"`
	UnbondingDelegations []UnbondingDelegation `json:"unbonding_delegations"`
	Redelegations        []Redelegation        `json:"redelegations"`
	Exported             bool                  `json:"exported"`
}

// Last validator power, needed for validator set update logic
type LastValidatorPower struct {
	Address sdk.ValAddress
	Power   int64
}

func NewGenesisState(pool Pool, params Params, validators []Validator, delegations []Delegation) GenesisState {
	return GenesisState{
		Pool:        pool,
		Params:      params,
		Validators:  validators,
		Delegations: delegations,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Pool:   InitialPool(),
		Params: DefaultParams(),
	}
}
