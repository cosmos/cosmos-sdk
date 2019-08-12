// DONTCOVER
// nolint
package v0_37

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_34"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_36"
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
		ConsPubKey              crypto.PubKey          `json:"consensus_pubkey" yaml:"consensus_pubkey"`
		Jailed                  bool                   `json:"jailed" yaml:"jailed"`
		Status                  sdk.BondStatus         `json:"status" yaml:"status"`
		Tokens                  sdk.Int                `json:"tokens" yaml:"tokens"`
		DelegatorShares         sdk.Dec                `json:"delegator_shares" yaml:"delegator_shares"`
		Description             Description            `json:"description" yaml:"description"`
		UnbondingHeight         int64                  `json:"unbonding_height" yaml:"unbonding_height"`
		UnbondingCompletionTime time.Time              `json:"unbonding_time" yaml:"unbonding_time"`
		Commission              v036staking.Commission `json:"commission" yaml:"commission"`
		MinSelfDelegation       sdk.Int                `json:"min_self_delegation" yaml:"min_self_delegation"`
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

// NewDescription creates a new Description object
func NewDescription(moniker, identity, website,
	securityContact, details string) Description {

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
