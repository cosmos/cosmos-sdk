package ibc

import (
	"reflect"

	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case OpenChannelMsg:
			return handleOpenChannelMsg(ctx, keeper, msg)
		case UpdateChannelMsg:
			return handleUpdateChannelMsg(ctx, keeper, msg)
		default:
			errMsg := "Unrecognized IBC Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleOpenChannelMsg(ctx sdk.Context, keeper Keeper, msg OpenChannelMsg) sdk.Result {
	_, err := keeper.getChannelCommitHeight(ctx, msg.SrcChain)
	if err == nil {
		return ErrChannelAlreadyOpened(msg.SrcChain).Result()
	}

	keeper.setChannelCommit(ctx, msg.SrcChain, msg.ROT.Height(), msg.ROT)

	return sdk.Result{}
}

func handleUpdateChannelMsg(ctx sdk.Context, keeper Keeper, msg UpdateChannelMsg) sdk.Result {
	height, err := keeper.getChannelCommitHeight(ctx, msg.SrcChain)
	if err != nil {
		return err.Result()
	}

	commit, ok := keeper.getChannelCommit(ctx, msg.SrcChain, height)
	if !ok {
		panic("Should not be happened")
	}

	cert := lite.NewDynamicCertifier(msg.SrcChain, commit.Validators, height)
	if err := cert.Update(msg.Commit); err != nil {
		return ErrUpdateCommitFailed(err).Result()
	}

	keeper.setChannelCommit(ctx, msg.SrcChain, msg.Commit.Height(), msg.Commit)

	return sdk.Result{}
}

type ReceiveHandler func(sdk.Context, Payload) (Payload, sdk.Error)

func (keeper Keeper) Receive(h ReceiveHandler, ctx sdk.Context, msg ReceiveMsg) sdk.Result {
	msg.Verify(ctx, keeper)

	packet := msg.Packet
	if packet.DestChain != ctx.ChainID() {
		return ErrChainMismatch().Result()
	}

	cctx, write := ctx.CacheContext()
	rec, err := h(cctx, packet.Payload)
	if rec != nil {
		keeper.sendReceipt(ctx, rec, packet.SrcChain)
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

func (keeper Keeper) Receipt(h ReceiptHandler, ctx sdk.Context, msg ReceiptMsg) sdk.Result {
	msg.Verify(ctx, keeper)

	h(ctx, msg.Payload)

	return sdk.Result{}
}
