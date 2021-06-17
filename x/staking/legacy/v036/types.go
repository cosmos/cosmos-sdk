// DONTCOVER
package v036

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v034"
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
		ConsPubKey              cryptotypes.PubKey      `json:"consensus_pubkey" yaml:"consensus_pubkey"`
		Jailed                  bool                    `json:"jailed" yaml:"jailed"`
		Status                  v034staking.BondStatus  `json:"status" yaml:"status"`
		Tokens                  sdk.Int                 `json:"tokens" yaml:"tokens"`
		DelegatorShares         sdk.Dec                 `json:"delegator_shares" yaml:"delegator_shares"`
		Description             v034staking.Description `json:"description" yaml:"description"`
		UnbondingHeight         int64                   `json:"unbonding_height" yaml:"unbonding_height"`
		UnbondingCompletionTime time.Time               `json:"unbonding_time" yaml:"unbonding_time"`
		Commission              Commission              `json:"commission" yaml:"commission"`
		MinSelfDelegation       sdk.Int                 `json:"min_self_delegation" yaml:"min_self_delegation"`
	}

	bechValidator struct {
		OperatorAddress         sdk.ValAddress          `json:"operator_address" yaml:"operator_address"`
		ConsPubKey              string                  `json:"consensus_pubkey" yaml:"consensus_pubkey"`
		Jailed                  bool                    `json:"jailed" yaml:"jailed"`
		Status                  v034staking.BondStatus  `json:"status" yaml:"status"`
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

func (v Validator) MarshalJSON() ([]byte, error) {
	bechConsPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, v.ConsPubKey)
	if err != nil {
		return nil, err
	}

	return legacy.Cdc.MarshalJSON(bechValidator{
		OperatorAddress:         v.OperatorAddress,
		ConsPubKey:              bechConsPubKey,
		Jailed:                  v.Jailed,
		Status:                  v.Status,
		Tokens:                  v.Tokens,
		DelegatorShares:         v.DelegatorShares,
		Description:             v.Description,
		UnbondingHeight:         v.UnbondingHeight,
		UnbondingCompletionTime: v.UnbondingCompletionTime,
		MinSelfDelegation:       v.MinSelfDelegation,
		Commission:              v.Commission,
	})
}

func (v *Validator) UnmarshalJSON(data []byte) error {
	bv := &bechValidator{}
	if err := legacy.Cdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	consPubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, bv.ConsPubKey)
	if err != nil {
		return err
	}
	*v = Validator{
		OperatorAddress:         bv.OperatorAddress,
		ConsPubKey:              consPubKey,
		Jailed:                  bv.Jailed,
		Tokens:                  bv.Tokens,
		Status:                  bv.Status,
		DelegatorShares:         bv.DelegatorShares,
		Description:             bv.Description,
		UnbondingHeight:         bv.UnbondingHeight,
		UnbondingCompletionTime: bv.UnbondingCompletionTime,
		Commission:              bv.Commission,
		MinSelfDelegation:       bv.MinSelfDelegation,
	}
	return nil
}
