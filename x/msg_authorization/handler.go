package msg_authorization

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/keeper"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgGrantAuthorization:
			return handleMsgGrantAuthorization(ctx, msg, k)
		case *types.MsgRevokeAuthorization:
			return handleMsgRevokeAuthorization(ctx, msg, k)
		case *types.MsgExecAuthorized:
			return handleMsgExecAuthorized(ctx, msg, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized authorization message type: %T", msg)
		}
	}
}

func handleMsgGrantAuthorization(ctx sdk.Context, msg *types.MsgGrantAuthorization, k keeper.Keeper) (*sdk.Result, error) {
	k.Grant(ctx, msg.Grantee, msg.Granter, msg.Authorization, msg.Expiration)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventGrantAuthorization,
			sdk.NewAttribute(types.AttributeKeyGrantType, msg.Authorization.MsgType()),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyGranterAddress, msg.Granter),
			sdk.NewAttribute(types.AttributeKeyGranteeAddress, msg.Grantee),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgRevokeAuthorization(ctx sdk.Context, msg *types.MsgRevokeAuthorization, k keeper.Keeper) (*sdk.Result, error) {
	err := k.Revoke(ctx, msg.Grantee, msg.Granter, msg.AuthorizationMsgType)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventRevokeAuthorization,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyGrantType, msg.AuthorizationMsgType),
			sdk.NewAttribute(types.AttributeKeyGranterAddress, msg.Granter),
			sdk.NewAttribute(types.AttributeKeyGranteeAddress, msg.Grantee),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgExecAuthorized(ctx sdk.Context, msg *types.MsgExecAuthorized, k keeper.Keeper) (*sdk.Result, error) {
	return k.DispatchActions(ctx, msg.Grantee, msg.Msgs)
}
