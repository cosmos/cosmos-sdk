package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

var _ types.QueryServer = ChangePubKeyKeeper{}

// Params returns parameters of changepubkey module
func (pk ChangePubKeyKeeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := pk.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}
