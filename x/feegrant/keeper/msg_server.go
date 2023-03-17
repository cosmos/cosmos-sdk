package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/feegrant"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the feegrant MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(k Keeper) feegrant.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

var _ feegrant.MsgServer = msgServer{}

// GrantAllowance grants an allowance from the granter's funds to be used by the grantee.
func (k msgServer) GrantAllowance(goCtx context.Context, msg *feegrant.MsgGrantAllowance) (*feegrant.MsgGrantAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	grantee, err := k.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}

	granter, err := k.addressCodec.StringToBytes(msg.Granter)
	if err != nil {
		return nil, err
	}

	if msg.Grantee == msg.Granter {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "cannot self-grant fee authorization")
	}

	allowance, err := msg.GetFeeAllowanceI()
	if err != nil {
		return nil, err
	}

	if err := allowance.ValidateBasic(); err != nil {
		return nil, err
	}

	err = k.Keeper.GrantAllowance(ctx, granter, grantee, allowance)
	if err != nil {
		return nil, err
	}

	return &feegrant.MsgGrantAllowanceResponse{}, nil
}

// RevokeAllowance revokes a fee allowance between a granter and grantee.
func (k msgServer) RevokeAllowance(goCtx context.Context, msg *feegrant.MsgRevokeAllowance) (*feegrant.MsgRevokeAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	grantee, err := k.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}

	granter, err := k.addressCodec.StringToBytes(msg.Granter)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.revokeAllowance(ctx, granter, grantee)
	if err != nil {
		return nil, err
	}

	if msg.Grantee == msg.Granter {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "addresses must be different")
	}

	return &feegrant.MsgRevokeAllowanceResponse{}, nil
}
