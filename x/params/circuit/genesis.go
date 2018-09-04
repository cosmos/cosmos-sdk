package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/space"
)

// GenesisState defines initial circuit brake msg types(usually empty)
type GenesisState struct {
	CircuitBrakeTypes []string `json:"circuit-brake-types"`
}

// InitGenesis spaces activated types to param space
func InitGenesis(ctx sdk.Context, space params.Space, data GenesisState) {
	for _, ty := range data.CircuitBrakeTypes {
		space.Set(ctx, CircuitBrakeKey(ty), true)
	}
}
