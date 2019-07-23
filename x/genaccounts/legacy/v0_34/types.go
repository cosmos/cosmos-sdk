// DONTCOVER
// nolint
package v0_34

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
	}

	GenesisState []GenesisAccount
)
