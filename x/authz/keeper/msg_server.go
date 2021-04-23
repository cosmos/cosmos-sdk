package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
)

var _ types.MsgServer = Keeper{}

// GrantAuthorization implements the MsgServer.Grant method.
func (k Keeper) Grant(goCtx context.Context, msg *types.MsgGrantRequest) (*types.MsgGrantResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		return nil, err
	}

	authorization := msg.GetAuthorization()
	if authorization == nil {
		return nil, sdkerrors.ErrUnpackAny.Wrap("Authorization is not present in the msg")
	}
	// If the granted service Msg doesn't exist, we throw an error.
	if k.router.Handler(authorization.MethodName()) == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "%s doesn't exist.", authorization.MethodName())
	}

	err = k.SaveGrant(ctx, grantee, granter, authorization, msg.Expiration)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrantResponse{}, nil
}

// RevokeAuthorization implements the MsgServer.Revoke method.
func (k Keeper) Revoke(goCtx context.Context, msg *types.MsgRevokeRequest) (*types.MsgRevokeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		return nil, err
	}

	err = k.DeleteGrant(ctx, grantee, granter, msg.MsgTypeUrl)
	if err != nil {
		return nil, err
	}

	return &types.MsgRevokeResponse{}, nil
}

// Exec implements the MsgServer.Exec method.
func (k Keeper) Exec(goCtx context.Context, msg *types.MsgExecRequest) (*types.MsgExecResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}
	msgs, err := msg.GetServiceMsgs()
	if err != nil {
		return nil, err
	}
	result, err := k.DispatchActions(ctx, grantee, msgs)
	if err != nil {
		return nil, err
	}
	return &types.MsgExecResponse{Result: result}, nil
}
