// DONTCOVER
// nolint
package v0_36

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_34"
)

const (
	ModuleName = "staking"
)

type (
	Commission struct {
		CommissionRates `json:"commission_rates" yaml:"commission_rates"`
		UpdateTime      time.Time `json:"update_time" yaml:"update_time"`
	}

	CommissionRates struct {
		Rate          sdk.Dec `json:"rate" yaml:"rate"`
		MaxRate       sdk.Dec `json:"max_rate" yaml:"max_rate"`
		MaxChangeRate sdk.Dec `json:"max_change_rate" yaml:"max_change_rate"`
	}

	Validator struct {
		OperatorAddress         sdk.ValAddress          `json:"operator_address" yaml:"operator_address"`
		ConsPubKey              crypto.PubKey           `json:"consensus_pubkey" yaml:"consensus_pubkey"`
		Jailed                  bool                    `json:"jailed" yaml:"jailed"`
		Status                  sdk.BondStatus          `json:"status" yaml:"status"`
		Tokens                  sdk.Int                 `json:"tokens" yaml:"tokens"`
		DelegatorShares         sdk.Dec                 `json:"delegator_shares" yaml:"delegator_shares"`
		Description             v034staking.Description `json:"description" yaml:"description"`
		UnbondingHeight         int64                   `json:"unbonding_height" yaml:"unbonding_height"`
		UnbondingCompletionTime time.Time               `json:"unbonding_time" yaml:"unbonding_time"`
		Commission              Commission              `json:"commission" yaml:"commission"`
		MinSelfDelegation       sdk.Int                 `json:"min_self_delegation" yaml:"min_self_delegation"`
	}

	Validators []Validator

	GenesisState struct {
		Params               v034staking.Params                `json:"params"`
		LastTotalPower       sdk.Int                           `json:"last_total_power"`
		LastValidatorPowers  []v034staking.LastValidatorPower  `json:"last_validator_powers"`
		Validators           Validators                        `json:"validators"`
		Delegations          v034staking.Delegations           `json:"delegations"`
		UnbondingDelegations []v034staking.UnbondingDelegation `json:"unbonding_delegations"`
		Redelegations        []v034staking.Redelegation        `json:"redelegations"`
		Exported             bool                              `json:"exported"`
	}
)

func NewGenesisState(
	params v034staking.Params, lastTotalPower sdk.Int, lastValPowers []v034staking.LastValidatorPower,
	validators Validators, delegations v034staking.Delegations,
	ubds []v034staking.UnbondingDelegation, reds []v034staking.Redelegation, exported bool,
) GenesisState {

	return GenesisState{
		Params:               params,
		LastTotalPower:       lastTotalPower,
		LastValidatorPowers:  lastValPowers,
		Validators:           validators,
		Delegations:          delegations,
		UnbondingDelegations: ubds,
		Redelegations:        reds,
		Exported:             exported,
	}
}
