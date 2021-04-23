package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the feegrant MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

var _ types.MsgServer = msgServer{}

// GrantFeeAllowance grants an allowance from the granter's funds to be used by the grantee.
func (k msgServer) GrantFeeAllowance(goCtx context.Context, msg *types.MsgGrantFeeAllowance) (*types.MsgGrantFeeAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}

	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		return nil, err
	}

	// Checking for duplicate entry
	if f, _ := k.Keeper.GetFeeAllowance(ctx, granter, grantee); f != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee allowance already exists")
	}

	allowance, err := msg.GetFeeAllowanceI()
	if err != nil {
		return nil, err
	}

	err = k.Keeper.GrantFeeAllowance(ctx, granter, grantee, allowance)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrantFeeAllowanceResponse{}, nil
}

// RevokeFeeAllowance revokes a fee allowance between a granter and grantee.
func (k msgServer) RevokeFeeAllowance(goCtx context.Context, msg *types.MsgRevokeFeeAllowance) (*types.MsgRevokeFeeAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}

	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.RevokeFeeAllowance(ctx, granter, grantee)
	if err != nil {
		return nil, err
	}

	return &types.MsgRevokeFeeAllowanceResponse{}, nil
}
