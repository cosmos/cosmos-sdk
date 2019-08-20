// DONTCOVER
// nolint
package v0_36

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "supply"

type (
	GenesisState struct {
		Supply sdk.Coins `json:"supply" yaml:"supply"`
	}
)

func EmptyGenesisState() GenesisState {
	return GenesisState{
		Supply: sdk.NewCoins(), // leave this empty as it's filled on initialization
	}
}
