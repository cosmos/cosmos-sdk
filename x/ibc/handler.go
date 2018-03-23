package ibc

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func NewHandler(keeper types.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case OpenChannelMsg:
			return handleOpenChannelMsg(ctx, keeper, msg)
		case UpdateChannelMsg:
			return handleUpdateChannelMsg(ctx, keeper, msg)
		case ReceiveMsg:
			return handleReceiveMsg(ctx, keeper, msg)
		default:
			errMsg := "Unrecognized IBC Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleOpenChannelMsg(ctx sdk.Context, keeper types.Keeper, msg OpenChannelMsg) sdk.Result {
	err := keeper.
}

func handleUpdateChannelMsg(ctx sdk.Context, keeper types.Keeper, msg UpdateChannelMsg) sdk.Result {

}

// ReceiveMsg adds coins to the destination address and creates an ingress IBC packet.
func handleReceiveMsg(ctx sdk.Context, keeper types.Keeper, msg ReceiveMsg) sdk.Result {
	err := keeper.Receive(ctx, msg.Packet, msg.Sequence)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
