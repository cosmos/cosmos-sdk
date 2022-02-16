// Package v038 is used for legacy migration scripts. Actual migration scripts
// for v038 have been removed, but the v039->v042 migration script still
// references types from this file, so we're keeping it for now.
// DONTCOVER
package v038

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v034"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v036"
)

const (
	ModuleName = "staking"
)

type (
	Description struct {
		Moniker         string `json:"moniker" yaml:"moniker"`
		Identity        string `json:"identity" yaml:"identity"`
		Website         string `json:"website" yaml:"website"`
		SecurityContact string `json:"security_contact" yaml:"security_contact"`
		Details         string `json:"details" yaml:"details"`
	}

	Validator struct {
		OperatorAddress         sdk.ValAddress         `json:"operator_address" yaml:"operator_address"`
		ConsPubKey              cryptotypes.PubKey     `json:"consensus_pubkey" yaml:"consensus_pubkey"`
		Jailed                  bool                   `json:"jailed" yaml:"jailed"`
		Status                  v034staking.BondStatus `json:"status" yaml:"status"`
		Tokens                  sdk.Int                `json:"tokens" yaml:"tokens"`
		DelegatorShares         sdk.Dec                `json:"delegator_shares" yaml:"delegator_shares"`
		Description             Description            `json:"description" yaml:"description"`
		UnbondingHeight         int64                  `json:"unbonding_height" yaml:"unbonding_height"`
		UnbondingCompletionTime time.Time              `json:"unbonding_time" yaml:"unbonding_time"`
		Commission              v036staking.Commission `json:"commission" yaml:"commission"`
		MinSelfDelegation       sdk.Int                `json:"min_self_delegation" yaml:"min_self_delegation"`
	}

	bechValidator struct {
		OperatorAddress         sdk.ValAddress         `json:"operator_address" yaml:"operator_address"`
		ConsPubKey              string                 `json:"consensus_pubkey" yaml:"consensus_pubkey"`
		Jailed                  bool                   `json:"jailed" yaml:"jailed"`
		Status                  v034staking.BondStatus `json:"status" yaml:"status"`
		Tokens                  sdk.Int                `json:"tokens" yaml:"tokens"`
		DelegatorShares         sdk.Dec                `json:"delegator_shares" yaml:"delegator_shares"`
		Description             Description            `json:"description" yaml:"description"`
		UnbondingHeight         int64                  `json:"unbonding_height" yaml:"unbonding_height"`
		UnbondingCompletionTime time.Time              `json:"unbonding_time" yaml:"unbonding_time"`
		Commission              v036staking.Commission `json:"commission" yaml:"commission"`
		MinSelfDelegation       sdk.Int                `json:"min_self_delegation" yaml:"min_self_delegation"`
	}

	Validators []Validator

	Params struct {
		UnbondingTime     time.Duration `json:"unbonding_time" yaml:"unbonding_time"`         // time duration of unbonding
		MaxValidators     uint16        `json:"max_validators" yaml:"max_validators"`         // maximum number of validators (max uint16 = 65535)
		MaxEntries        uint16        `json:"max_entries" yaml:"max_entries"`               // max entries for either unbonding delegation or redelegation (per pair/trio)
		HistoricalEntries uint16        `json:"historical_entries" yaml:"historical_entries"` // number of historical entries to persist
		BondDenom         string        `json:"bond_denom" yaml:"bond_denom"`                 // bondable coin denomination
	}

	GenesisState struct {
		Params               Params                            `json:"params"`
		LastTotalPower       sdk.Int                           `json:"last_total_power"`
		LastValidatorPowers  []v034staking.LastValidatorPower  `json:"last_validator_powers"`
		Validators           Validators                        `json:"validators"`
		Delegations          v034staking.Delegations           `json:"delegations"`
		UnbondingDelegations []v034staking.UnbondingDelegation `json:"unbonding_delegations"`
		Redelegations        []v034staking.Redelegation        `json:"redelegations"`
		Exported             bool                              `json:"exported"`
	}
)

// NewDescription creates a new Description object
func NewDescription(moniker, identity, website, securityContact, details string) Description {
	return Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params v034staking.Params, lastTotalPower sdk.Int, lastValPowers []v034staking.LastValidatorPower,
	validators Validators, delegations v034staking.Delegations,
	ubds []v034staking.UnbondingDelegation, reds []v034staking.Redelegation, exported bool,
) GenesisState {

	return GenesisState{
		Params: Params{
			UnbondingTime:     params.UnbondingTime,
			MaxValidators:     params.MaxValidators,
			MaxEntries:        params.MaxEntries,
			BondDenom:         params.BondDenom,
			HistoricalEntries: 0,
		},
		LastTotalPower:       lastTotalPower,
		LastValidatorPowers:  lastValPowers,
		Validators:           validators,
		Delegations:          delegations,
		UnbondingDelegations: ubds,
		Redelegations:        reds,
		Exported:             exported,
	}
}

// MarshalJSON marshals the validator to JSON using Bech32
func (v Validator) MarshalJSON() ([]byte, error) {
	bechConsPubKey, err := legacybech32.MarshalPubKey(legacybech32.ConsPK, v.ConsPubKey)
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

// UnmarshalJSON unmarshals the validator from JSON using Bech32
func (v *Validator) UnmarshalJSON(data []byte) error {
	bv := &bechValidator{}
	if err := legacy.Cdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	consPubKey, err := legacybech32.UnmarshalPubKey(legacybech32.ConsPK, bv.ConsPubKey)
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
