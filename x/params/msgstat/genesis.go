package msgstat

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/store"
)

// ActivatedParamKey - paramstore key for msg type activation
const (
	DefaultParamSpace = "MsgStatus"
)

// GenesisState defines initial activated msg types
type GenesisState struct {
	ActivatedTypes []string `json:"activated-types"`
}

// InitGenesis stores activated types to param store
func InitGenesis(ctx sdk.Context, store params.Store, data GenesisState) {
	for _, ty := range data.ActivatedTypes {
		store.Set(ctx, params.NewKey(ty), true)
	}
}
