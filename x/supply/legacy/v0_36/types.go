// DONTCOVER
// nolint
package v0_36

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "supply"

type (
	Supply struct {
		Total sdk.Coins `json:"total"`
	}

	GenesisState struct {
		Supply Supply `json:"supply"`
	}
)

func EmptyGenesisState() GenesisState {
	return GenesisState{
		Supply: Supply{
			Total: sdk.NewCoins(), // leave this empty as it's filled on initialization
		},
	}
}
