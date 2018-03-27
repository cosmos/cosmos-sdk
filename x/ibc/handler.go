package ibc

import (
	"fmt"

	"reflect"

	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func NewHandler(keeper keeper) sdk.Handler {
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

func handleOpenChannelMsg(ctx sdk.Context, keeper keeper, msg OpenChannelMsg) sdk.Result {
	_, err := keeper.getChannelCommitHeight(ctx, msg.SrcChain)
	if err == nil {
		return types.ErrChannelAlreadyOpened(msg.SrcChain).Result()
	}

	keeper.setChannelCommit(ctx, msg.SrcChain, msg.ROT.Height(), msg.ROT)

	return sdk.Result{}
}

func handleUpdateChannelMsg(ctx sdk.Context, keeper keeper, msg UpdateChannelMsg) sdk.Result {
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
		return types.ErrUpdateCommitFailed(err).Result()
	}

	keeper.setChannelCommit(ctx, msg.SrcChain, msg.Commit.Height(), msg.Commit)

	return sdk.Result{}
}

// ReceiveMsg adds coins to the destination address and creates an ingress IBC packet.
func handleReceiveMsg(ctx sdk.Context, keeper keeper, msg ReceiveMsg) sdk.Result {
	expected := keeper.getIngressSequence(ctx, msg.SrcChain)
	seq := msg.Sequence
	if seq != expected {
		return types.ErrInvalidSequence().Result()
	}

	keeper.setIngressSequence(ctx, msg.SrcChain, seq+1)

	commit, ok := keeper.getChannelCommit(ctx, msg.SrcChain, msg.Height)
	if !ok {
		return types.ErrNoCommitFound().Result()
	}

	key := []byte(fmt.Sprintf("ibc/%s", EgressKey(ctx.ChainID(), msg.Sequence)))
	value, rawerr := keeper.cdc.MarshalBinary(msg.Packet) // better way to do this?
	if rawerr != nil {
		return types.ErrInvalidPacket(rawerr).Result()
	}

	if rawerr = msg.Proof.Verify(key, value, commit.Commit.Header.AppHash); rawerr != nil {
		return types.ErrInvalidPacket(rawerr).Result()
	}

	err := keeper.Receive(ctx, msg.Packet)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
