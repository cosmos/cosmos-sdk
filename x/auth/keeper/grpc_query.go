package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ types.QueryServer = AccountKeeper{}

// Account returns account details based on address
func (k AccountKeeper) Account(c context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	account := k.GetAccount(ctx, req.Address)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", req.Address)
	}

	return &types.QueryAccountResponse{Account: account}, nil
}

// Parameters returns parameters of auth module
func (k AccountKeeper) Parameters(c context.Context, req *types.QueryParametersRequest) (*types.QueryParametersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParametersResponse{Params: params}, nil
}
