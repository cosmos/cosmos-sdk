package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type GenesisState struct {
	MsgTypes []string `json:"msg-types"`
	MsgNames []string `json:"msg-names"`
}

func InitGenesis(ctx sdk.Context, space params.Subspace, data GenesisState) {
	for _, ty := range data.MsgTypes {
		space.SetWithSubkey(ctx, MsgTypeKey, []byte(ty), true)
	}
	for _, name := range data.MsgNames {
		space.SetWithSubkey(ctx, MsgNameKey, []byte(name), true)
	}
}
