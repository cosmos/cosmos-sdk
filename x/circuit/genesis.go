package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisState struct {
	MsgTypes []string `json:"msg-types"`
	MsgNames []string `json:"msg-names"`
}

func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for _, ty := range data.MsgTypes {
		k.space.SetWithSubkey(ctx, MsgTypeKey, []byte(ty), true)
	}
	for _, name := range data.MsgNames {
		k.space.SetWithSubkey(ctx, MsgNameKey, []byte(name), true)
	}
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		MsgTypes: nil,
		MsgNames: nil,
	}
}
