// DONTCOVER
package v036

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "accounts"
)

type (
	GenesisAccount struct {
		Address       sdk.AccAddress `json:"address" yaml:"address"`
		Coins         sdk.Coins      `json:"coins" yaml:"coins"`
		Sequence      uint64         `json:"sequence_number" yaml:"sequence_number"`
		AccountNumber uint64         `json:"account_number" yaml:"account_number"`

		OriginalVesting  sdk.Coins `json:"original_vesting" yaml:"original_vesting"`
		DelegatedFree    sdk.Coins `json:"delegated_free" yaml:"delegated_free"`
		DelegatedVesting sdk.Coins `json:"delegated_vesting" yaml:"delegated_vesting"`
		StartTime        int64     `json:"start_time" yaml:"start_time"`
		EndTime          int64     `json:"end_time" yaml:"end_time"`

		ModuleName        string   `json:"module_name" yaml:"module_name"`
		ModulePermissions []string `json:"module_permissions" yaml:"module_permissions"`
	}

	GenesisState []GenesisAccount
)

// NewGenesisAccount creates a new GenesisAccount object
func NewGenesisAccount(
	address sdk.AccAddress, coins sdk.Coins, sequence uint64,
	vestingAmount, delFree, delVesting sdk.Coins, vestingStartTime, vestingEndTime int64,
	module string, permissions []string,
) GenesisAccount {

	return GenesisAccount{
		Address:           address,
		Coins:             coins,
		Sequence:          sequence,
		AccountNumber:     0, // ignored set by the account keeper during InitGenesis
		OriginalVesting:   vestingAmount,
		DelegatedFree:     delFree,
		DelegatedVesting:  delVesting,
		StartTime:         vestingStartTime,
		EndTime:           vestingEndTime,
		ModuleName:        module,
		ModulePermissions: permissions,
	}
}
