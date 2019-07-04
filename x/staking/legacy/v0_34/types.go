// DONTCOVER
// nolint
package v0_34

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "staking"
)

type (
	Pool struct {
		NotBondedTokens sdk.Int `json:"not_bonded_tokens"`
		BondedTokens    sdk.Int `json:"bonded_tokens"`
	}

	Params struct {
		UnbondingTime time.Duration `json:"unbonding_time"`
		MaxValidators uint16        `json:"max_validators"`
		MaxEntries    uint16        `json:"max_entries"`
		BondDenom     string        `json:"bond_denom"`
	}

	LastValidatorPower struct {
		Address sdk.ValAddress `json:"address"`
		Power   int64          `json:"power"`
	}

	Description struct {
		Moniker  string `json:"moniker"`
		Identity string `json:"identity"`
		Website  string `json:"website"`
		Details  string `json:"details"`
	}

	Commission struct {
		CommissionRates CommissionRates `json:"commission_rates"`
		UpdateTime time.Time `json:"update_time"`
	}

	CommissionRates struct {
		Rate          sdk.Dec `json:"rate"`
		MaxRate       sdk.Dec `json:"max_rate"`
		MaxChangeRate sdk.Dec `json:"max_change_rate"`
	}

	Validator struct {
		OperatorAddress         sdk.ValAddress `json:"operator_address"`
		ConsPubKey              crypto.PubKey  `json:"consensus_pubkey"`
		Jailed                  bool           `json:"jailed"`
		Status                  sdk.BondStatus `json:"status"`
		Tokens                  sdk.Int        `json:"tokens"`
		DelegatorShares         sdk.Dec        `json:"delegator_shares"`
		Description             Description    `json:"description"`
		UnbondingHeight         int64          `json:"unbonding_height"`
		UnbondingCompletionTime time.Time      `json:"unbonding_time"`
		Commission              Commission     `json:"commission"`
		MinSelfDelegation       sdk.Int        `json:"min_self_delegation"`
	}

	Validators []Validator

	Delegation struct {
		DelegatorAddress sdk.AccAddress `json:"delegator_address"`
		ValidatorAddress sdk.ValAddress `json:"validator_address"`
		Shares           sdk.Dec        `json:"shares"`
	}

	Delegations []Delegation

	UnbondingDelegationEntry struct {
		CreationHeight int64     `json:"creation_height"`
		CompletionTime time.Time `json:"completion_time"`
		InitialBalance sdk.Int   `json:"initial_balance"`
		Balance        sdk.Int   `json:"balance"`
	}

	UnbondingDelegation struct {
		DelegatorAddress sdk.AccAddress             `json:"delegator_address"`
		ValidatorAddress sdk.ValAddress             `json:"validator_address"`
		Entries          []UnbondingDelegationEntry `json:"entries"`
	}

	RedelegationEntry struct {
		CreationHeight int64     `json:"creation_height"`
		CompletionTime time.Time `json:"completion_time"`
		InitialBalance sdk.Int   `json:"initial_balance"`
		SharesDst      sdk.Dec   `json:"shares_dst"`
	}

	Redelegation struct {
		DelegatorAddress    sdk.AccAddress      `json:"delegator_address"`
		ValidatorSrcAddress sdk.ValAddress      `json:"validator_src_address"`
		ValidatorDstAddress sdk.ValAddress      `json:"validator_dst_address"`
		Entries             []RedelegationEntry `json:"entries"`
	}

	GenesisState struct {
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
)
