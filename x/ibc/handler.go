package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(ibcm IBCMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case IBCTransferMsg:
			return handleIBCTransferMsg(ctx, ibcm, msg)
		case IBCReceiveMsg:
			return handleIBCReceiveMsg(ctx, ibcm, msg)
		}
	}
}

func handleIBCTransferMsg(ctx sdk.Context, ibcm IBCMapper, msg IBCTransferMsg) sdk.Result {
	ibcm.PushPacket(ctx, msg.IBCPacket)
	return sdk.Result{}
}

func handleIBCReceiveMsg(ctx sdk.Context, ibcm IBCMapper, msg IBCReceiveMsg) sdk.Result {
	seq := ibc.IngressSequence(packet.SrcChain)
}
