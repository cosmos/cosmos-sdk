// DONTCOVER
// nolint
package v0_36

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "accounts"
)

type (
	GenesisAccount struct {
		Address       sdk.AccAddress `json:"address"`
		Coins         sdk.Coins      `json:"coins"`
		Sequence      uint64         `json:"sequence_number"`
		AccountNumber uint64         `json:"account_number"`

		OriginalVesting  sdk.Coins `json:"original_vesting"`
		DelegatedFree    sdk.Coins `json:"delegated_free"`
		DelegatedVesting sdk.Coins `json:"delegated_vesting"`
		StartTime        int64     `json:"start_time"`
		EndTime          int64     `json:"end_time"`

		ModuleName       string `json:"module_name"`
		ModulePermission string `json:"module_permission"`
	}

	Description struct {
		Moniker  string `json:"moniker"`
		Identity string `json:"identity"`
		Website  string `json:"website"`
		Details  string `json:"details"`
	}

	Commission struct {
		CommissionRates CommissionRates `json:"commission_rates"`
		UpdateTime      time.Time       `json:"update_time"`
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

	GenesisState []GenesisAccount
)


// NewGenesisAccount creates a new GenesisAccount object
func NewGenesisAccount(address sdk.AccAddress, coins sdk.Coins, sequence uint64,
vestingAmount, delFree, delVesting sdk.Coins,
vestingStartTime, vestingEndTime int64,
module, permission string) GenesisAccount {
	
	return GenesisAccount{
		Address:          address,
		Coins:            coins,
		Sequence:         sequence,
		AccountNumber:    0, // ignored set by the account keeper during InitGenesis
		OriginalVesting:  vestingAmount,
		DelegatedFree:   	delFree,
		DelegatedVesting: delVesting,
		StartTime:        vestingStartTime,
		EndTime:          vestingEndTime,
		ModuleName:       module,
		ModulePermission: permission,
	}
}