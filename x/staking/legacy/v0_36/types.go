// DONTCOVER
// nolint
package v0_36

import (
	"time"
	
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "staking"
)

type (
	Params struct {
		MaxMemoCharacters      uint64 `json:"max_memo_characters"`
		TxSigLimit             uint64 `json:"tx_sig_limit"`
		TxSizeCostPerByte      uint64 `json:"tx_size_cost_per_byte"`
		SigVerifyCostED25519   uint64 `json:"sig_verify_cost_ed25519"`
		SigVerifyCostSecp256k1 uint64 `json:"sig_verify_cost_secp256k1"`
	}

	LastValidatorPower struct {
		Address sdk.ValAddress `json:"address"`
		Power   int64 `json:"power"`
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

func NewGenesisState(params Params, lastTotalPower sdk.Int,
	validators Validators, delegations Delegations, ubds []UnbondingDelegation,
	reds []Redelegation, exported bool) GenesisState {
	return GenesisState{
		Params:      params,
		LastTotalPower:       sdk.Int,
		LastValidatorPowers:  []LastValidatorPower,
		Validators:  validators,
		Delegations: delegations,
		UnbondingDelegations: ubds,
		Redelegations: reds,
		Exported: exported,
	}
}