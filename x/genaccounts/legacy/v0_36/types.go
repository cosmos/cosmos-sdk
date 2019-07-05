// DONTCOVER
// nolint
package v0_36

import (
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

	GenesisState []GenesisAccount
)

// NewGenesisAccount creates a new GenesisAccount object
func NewGenesisAccount(
	address sdk.AccAddress, coins sdk.Coins, sequence uint64,
	vestingAmount, delFree, delVesting sdk.Coins, vestingStartTime, vestingEndTime int64,
	module, permission string,
) GenesisAccount {

	return GenesisAccount{
		Address:          address,
		Coins:            coins,
		Sequence:         sequence,
		AccountNumber:    0, // ignored set by the account keeper during InitGenesis
		OriginalVesting:  vestingAmount,
		DelegatedFree:    delFree,
		DelegatedVesting: delVesting,
		StartTime:        vestingStartTime,
		EndTime:          vestingEndTime,
		ModuleName:       module,
		ModulePermission: permission,
	}
}
