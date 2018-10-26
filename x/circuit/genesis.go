package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Genesis state for circuit breaker
// Defines initial circuit break msg types and names
type GenesisState struct {
	MsgRoutes []string `json:"msg-routes"`
	MsgTypes  []string `json:"msg-types"`
}

// InitGenesis stores GenesisState into paramstore
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for _, route := range data.MsgRoutes {
		k.space.SetWithSubkey(ctx, MsgRouteKey, []byte(route), true)
	}
	for _, ty := range data.MsgTypes {
		k.space.SetWithSubkey(ctx, MsgTypeKey, []byte(ty), true)
	}
}

// No msg is break by default
func DefaultGenesisState() GenesisState {
	return GenesisState{
		MsgRoutes: nil,
		MsgTypes:  nil,
	}
}
