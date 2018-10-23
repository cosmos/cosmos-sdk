package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Genesis state for circuit breaker
// Defines initial circuit break msg types and names
type GenesisState struct {
	MsgTypes []string `json:"msg-types"`
	MsgNames []string `json:"msg-names"`
}

// InitGenesis stores GenesisState into paramstore
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for _, ty := range data.MsgTypes {
		k.space.SetWithSubkey(ctx, MsgTypeKey, []byte(ty), true)
	}
	for _, name := range data.MsgNames {
		k.space.SetWithSubkey(ctx, MsgNameKey, []byte(name), true)
	}
}

// No msg is break by default
func DefaultGenesisState() GenesisState {
	return GenesisState{
		MsgTypes: nil,
		MsgNames: nil,
	}
}
