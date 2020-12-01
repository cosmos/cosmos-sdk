package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

func (k msgServer) GrantFeeAllowance(goCtx context.Context, msg *types.MsgGrantFeeAllowance) (*types.MsgGrantFeeAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	grantee, err := sdk.AccAddressFromBech32(msg.Grantee.String())
	if err != nil {
		return nil, err
	}

	granter, err := sdk.AccAddressFromBech32(msg.Granter.String())
	if err != nil {
		return nil, err
	}

	feegrant := types.FeeAllowanceGrant{
		Grantee:   grantee,
		Granter:   granter,
		Allowance: msg.Allowance,
	}

	k.Keeper.GrantFeeAllowance(ctx, feegrant)

	return &types.MsgGrantFeeAllowanceResponse{}, nil
}

func (k msgServer) RevokeFeeAllowance(goCtx context.Context, msg *types.MsgRevokeFeeAllowance) (*types.MsgRevokeFeeAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	grantee, err := sdk.AccAddressFromBech32(msg.Grantee.String())
	if err != nil {
		return nil, err
	}

	granter, err := sdk.AccAddressFromBech32(msg.Granter.String())
	if err != nil {
		return nil, err
	}

	k.Keeper.RevokeFeeAllowance(ctx, granter, grantee)

	return &types.MsgRevokeFeeAllowanceResponse{}, nil
}
