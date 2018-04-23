package ibc

import (
	"reflect"

	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(keeper keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case OpenChannelMsg:
			return handleOpenChannelMsg(ctx, keeper, msg)
		case UpdateChannelMsg:
			return handleUpdateChannelMsg(ctx, keeper, msg)
		case ReceiveCleanupMsg:
			return handleReceiveCleanupMsg(ctx, keeper, msg)
		case ReceiptCleanupMsg:
			return handleReceiptCleanupMsg(ctx, keeper, msg)
		default:
			errMsg := "Unrecognized IBC Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleOpenChannelMsg(ctx sdk.Context, keeper keeper, msg OpenChannelMsg) sdk.Result {
	_, err := keeper.getCommitHeight(ctx, msg.SrcChain)
	if err == nil {
		return ErrChannelAlreadyOpened(msg.SrcChain).Result()
	}

	keeper.setCommit(ctx, msg.SrcChain, msg.ROT.Height(), msg.ROT)

	return sdk.Result{}
}

func handleUpdateChannelMsg(ctx sdk.Context, keeper keeper, msg UpdateChannelMsg) sdk.Result {
	height, err := keeper.getCommitHeight(ctx, msg.SrcChain)
	if err != nil {
		return err.Result()
	}

	commit, ok := keeper.getCommit(ctx, msg.SrcChain, height)
	if !ok {
		panic("Should not be happened")
	}

	cert := lite.NewDynamicCertifier(msg.SrcChain, commit.Validators, height)
	if err := cert.Update(msg.Commit); err != nil {
		return ErrUpdateCommitFailed(err).Result()
	}

	keeper.setCommit(ctx, msg.SrcChain, msg.Commit.Height(), msg.Commit)

	return sdk.Result{}
}

type ReceiveHandler func(sdk.Context, Payload) (Payload, sdk.Error)

func (channel Channel) Receive(h ReceiveHandler, ctx sdk.Context, msg ReceiveMsg) sdk.Result {
	keeper := channel.keeper

	if err := msg.Verify(ctx, keeper); err != nil {
		return err.Result()
	}

	packet := msg.Packet
	if packet.DestChain != ctx.ChainID() {
		return ErrChainMismatch().Result()
	}

	cctx, write := ctx.CacheContext()
	rec, err := h(cctx, packet.Payload)
	if rec != nil {
		if rec.Type() != channel.name {
			return ErrUnauthorizedSendReceipt().Result()
		}

		recPacket := Packet{
			Payload:   rec,
			SrcChain:  ctx.ChainID(),
			DestChain: packet.SrcChain,
		}

		keeper.receipt.Push(ctx, recPacket)
	}
	if err != nil {
		return sdk.Result{
			Code: sdk.CodeOK,
			Log:  err.ABCILog(),
		}
	}
	write()

	return sdk.Result{}
}

type ReceiptHandler func(sdk.Context, Payload)

func (channel Channel) Receipt(h ReceiptHandler, ctx sdk.Context, msg ReceiptMsg) sdk.Result {
	if err := msg.Verify(ctx, channel.keeper); err != nil {
		return err.Result()
	}

	h(ctx, msg.Payload)

	return sdk.Result{}
}

func handleReceiveCleanupMsg(ctx sdk.Context, keeper keeper, msg ReceiveCleanupMsg) sdk.Result {
	receive := keeper.receive

	if err := msg.Verify(ctx, receive, msg.SrcChain, msg.Sequence); err != nil {
		return err.Result()
	}

	// TODO: cleanup

	return sdk.Result{}
}

func handleReceiptCleanupMsg(ctx sdk.Context, keeper keeper, msg ReceiptCleanupMsg) sdk.Result {
	receipt := keeper.receipt

	if err := msg.Verify(ctx, receipt, msg.SrcChain, msg.Sequence); err != nil {
		return err.Result()
	}

	// TODO: cleanup

	return sdk.Result{}
}
