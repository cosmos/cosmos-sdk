package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(ibcm Mapper, ck BankKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case IBCTransferMsg:
			return handleIBCTransferMsg(ctx, ibcm, ck, msg)
		case IBCReceiveMsg:
			return handleIBCReceiveMsg(ctx, ibcm, ck, msg)
		default:
			errMsg := "Unrecognized IBC Msg type: " + msg.Type()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// IBCTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleIBCTransferMsg(ctx sdk.Context, ibcm Mapper, ck BankKeeper, msg IBCTransferMsg) sdk.Result {
	packet := msg.IBCPacket

	_, _, err := ck.SubtractCoins(ctx, packet.SrcAddr, packet.Coins)
	if err != nil {
		return err.Result()
	}

	err = ibcm.PostIBCPacket(ctx, packet)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}

// IBCReceiveMsg adds coins to the destination address and creates an ingress IBC packet.
func handleIBCReceiveMsg(ctx sdk.Context, ibcm Mapper, ck BankKeeper, msg IBCReceiveMsg) sdk.Result {
	packet := msg.IBCPacket

	seq := ibcm.GetIngressSequence(ctx, packet.SrcChain)
	if msg.Sequence != seq {
		return ErrInvalidSequence(ibcm.codespace).Result()
	}

	_, _, err := ck.AddCoins(ctx, packet.DestAddr, packet.Coins)
	if err != nil {
		return err.Result()
	}

	ibcm.SetIngressSequence(ctx, packet.SrcChain, seq+1)

	return sdk.Result{}
}
