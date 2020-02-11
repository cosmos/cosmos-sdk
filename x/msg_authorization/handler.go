package msg_authorization

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgGrantAuthorization:
			return handleMsgGrantAuthorization(ctx, msg, k)
		case MsgRevokeAuthorization:
			return handleMsgRevokeAuthorization(ctx, msg, k)
		case MsgExecDelegated:
			return handleMsgExecDelegated(ctx, msg, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized authorization message type: %T", msg)
		}
	}
}

func handleMsgGrantAuthorization(ctx sdk.Context, msg MsgGrantAuthorization, k Keeper) (*sdk.Result, error) {
	granted := k.Grant(ctx, msg.Grantee, msg.Granter, msg.Authorization, msg.Expiration)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyGranterAddress, msg.Granter.String()),
			sdk.NewAttribute(types.AttributeKeyGranteeAddress, msg.Grantee.String()),
		),
	)

	if granted {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventGrantAuthorization,
				sdk.NewAttribute(types.AttributeGranttype, msg.Authorization.MsgType()),
				sdk.NewAttribute(types.AttributeKeyGranterAddress, msg.Granter.String()),
				sdk.NewAttribute(types.AttributeKeyGranteeAddress, msg.Grantee.String()),
			),
		)
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgRevokeAuthorization(ctx sdk.Context, msg MsgRevokeAuthorization, k Keeper) (*sdk.Result, error) {
	revoked := k.Revoke(ctx, msg.Grantee, msg.Granter, msg.AuthorizationMsgType)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyGranterAddress, msg.Granter.String()),
			sdk.NewAttribute(types.AttributeKeyGranteeAddress, msg.Grantee.String()),
		),
	)

	if revoked {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventGrantAuthorization,
				sdk.NewAttribute(types.AttributeGranttype, msg.AuthorizationMsgType),
				sdk.NewAttribute(types.AttributeKeyGranterAddress, msg.Granter.String()),
				sdk.NewAttribute(types.AttributeKeyGranteeAddress, msg.Grantee.String()),
			),
		)
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgExecDelegated(ctx sdk.Context, msg MsgExecDelegated, k Keeper) (*sdk.Result, error) {
	return k.DispatchActions(ctx, msg.Grantee, msg.Msgs)
}
