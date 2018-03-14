package ibc

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(ibcm IBCMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case IBCTransferMsg:
			return handleIBCTransferMsg(ctx, ibcm, msg)
		case IBCReceiveMsg:
			return handleIBCReceiveMsg(ctx, ibcm, msg)
		default:
			errMsg := "Unrecognized IBC Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleIBCTransferMsg(ctx sdk.Context, ibcm IBCMapper, msg IBCTransferMsg) sdk.Result {
	ibcm.PushPacket(ctx, msg.IBCPacket)
	return sdk.Result{}
}

func handleIBCReceiveMsg(ctx sdk.Context, ibcm IBCMapper, msg IBCReceiveMsg) sdk.Result {
	packet := msg.IBCPacket
	seq := ibcm.GetIngressSequence(ctx, packet.SrcChain)
	if msg.Sequence != seq {
		return sdk.Result{} // error
	}
	ibcm.SetIngressSequence(ctx, packet.SrcChain, seq+1)

	// handle packet
	// packet.Handle(ctx)...

	return sdk.Result{}
}
