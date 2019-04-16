package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(ibcm Mapper, ck BankKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgIBCTransfer:
			return handleIBCTransferMsg(ctx, ibcm, ck, msg)

		case MsgIBCReceive:
			return handleIBCReceiveMsg(ctx, ibcm, ck, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// MsgIBCTransfer deducts coins from the account and creates an egress IBC packet.
func handleIBCTransferMsg(ctx sdk.Context, ibcm Mapper, ck BankKeeper, msg MsgIBCTransfer) sdk.Result {
	packet := msg.IBCPacket

	_, err := ck.SubtractCoins(ctx, packet.SrcAddr, packet.Coins)
	if err != nil {
		return err.Result()
	}

	err = ibcm.PostIBCPacket(ctx, packet)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}

// MsgIBCReceive adds coins to the destination address and creates an ingress IBC packet.
func handleIBCReceiveMsg(ctx sdk.Context, ibcm Mapper, ck BankKeeper, msg MsgIBCReceive) sdk.Result {
	packet := msg.IBCPacket

	seq := ibcm.GetIngressSequence(ctx, packet.SrcChain)
	if msg.Sequence != seq {
		return ErrInvalidSequence(ibcm.codespace).Result()
	}

	// XXX Check that packet.Coins is valid and positive (nonzero)
	_, err := ck.AddCoins(ctx, packet.DestAddr, packet.Coins)
	if err != nil {
		return err.Result()
	}

	ibcm.SetIngressSequence(ctx, packet.SrcChain, seq+1)

	return sdk.Result{}
}
