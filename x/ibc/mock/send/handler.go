package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/mock/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgSequence:
			return handleMsgSequence(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("21345").Result()
		}
	}
}

func handleMsgSequence(ctx sdk.Context, k Keeper, msg types.MsgSequence) (res sdk.Result) {
	err := k.UpdateSequence(ctx, msg.ChannelID, msg.Sequence)
	if err != nil {
		return err.Result()
	}
	return res
}
