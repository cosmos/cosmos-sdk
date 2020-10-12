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
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}

	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		return nil, err
	}

	authorization := msg.Authorization.GetCachedValue().(types.Authorization)

	k.Grant(ctx, grantee, granter, authorization, msg.Expiration)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventGrantAuthorization,
			sdk.NewAttribute(types.AttributeKeyGrantType, authorization.MsgType()),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyGranterAddress, msg.Granter),
			sdk.NewAttribute(types.AttributeKeyGranteeAddress, msg.Grantee),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgRevokeAuthorization(ctx sdk.Context, msg *types.MsgRevokeAuthorization, k keeper.Keeper) (*sdk.Result, error) {
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}

	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		return nil, err
	}
	err = k.Revoke(ctx, grantee, granter, msg.AuthorizationMsgType)
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
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}

	msgs, err := msg.GetMsgs()
	if err != nil {
		return nil, err
	}
	return k.DispatchActions(ctx, grantee, msgs)
}
