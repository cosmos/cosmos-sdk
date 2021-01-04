package changepubkey

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	changepubkeykeeper "github.com/cosmos/cosmos-sdk/x/auth/changepubkey/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

type msgServer struct {
	keeper.AccountKeeper
	historyKeeper changepubkeykeeper.Keeper
}

// NewMsgServerImpl returns an implementation of the changepubkey MsgServer interface,
// wrapping the corresponding AccountKeeper.
func NewMsgServerImpl(k keeper.AccountKeeper, historyKeeper changepubkeykeeper.Keeper) types.MsgServer {
	return &msgServer{k, historyKeeper}
}

var _ types.MsgServer = msgServer{}

func (s msgServer) ChangePubKey(goCtx context.Context, msg *types.MsgChangePubKey) (*types.MsgChangePubKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := s.AccountKeeper

	acc := ak.GetAccount(ctx, msg.Address)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s does not exist", msg.Address)
	}
	// TODO: what time would be good to be put here?
	s.historyKeeper.StoreLastPubKey(ctx, msg.Address, ctx.BlockTime(), acc.GetPubKey())

	acc.SetPubKey(msg.GetPubKey())
	ak.SetAccount(ctx, acc)

	// handle additional fee logic inside MsgChangePubKey handler
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signers should exist")
	}

	authParams := ak.GetParams(ctx)
	if !authParams.EnableChangePubKey {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "change pubkey param is disabled")
	}
	ctx.GasMeter().ConsumeGas(authParams.PubKeyChangeCost, "pubkey change fee")

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgChangePubKeyResponse{}, nil
}
