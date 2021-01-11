package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/x/fee_grant/types"
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
	f := k.Keeper.GetFeeAllowance(ctx, granter, grantee)
	if f != nil {
		return nil, status.Errorf(codes.AlreadyExists, "grant already exist")
	}

	k.Keeper.GrantFeeAllowance(ctx, granter, grantee, msg.GetFeeAllowanceI())

	return &types.MsgGrantFeeAllowanceResponse{}, nil
}

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
