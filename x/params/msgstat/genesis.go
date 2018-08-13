package msgstat

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/space"
)

// ActivatedParamKey - paramspace key for msg type activation
const (
	DefaultParamSpace = "MsgStatus"
)

// GenesisState defines initial activated msg types
type GenesisState struct {
	ActivatedTypes []string `json:"activated-types"`
}

// InitGenesis spaces activated types to param space
func InitGenesis(ctx sdk.Context, space params.Space, data GenesisState) {
	for _, ty := range data.ActivatedTypes {
		space.Set(ctx, params.NewKey(ty), true)
	}
}
