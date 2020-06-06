package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ types.QueryServer = BaseKeeper{}

// Balance implements the Query/Balance gRPC method
func (q BaseKeeper) Balance(c context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(req.Address) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address")
	}

	if req.Denom == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(c)
	balance := q.GetBalance(ctx, req.Address, req.Denom)

	return &types.QueryBalanceResponse{Balance: &balance}, nil
}

// AllBalances implements the Query/AllBalances gRPC method
func (q BaseKeeper) AllBalances(c context.Context, req *types.QueryAllBalancesRequest) (*types.QueryAllBalancesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(req.Address) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(c)
	balances := q.GetAllBalances(ctx, req.Address)

	return &types.QueryAllBalancesResponse{Balances: balances}, nil
}

// TotalSupply implements the Query/TotalSupply gRPC method
func (q BaseKeeper) TotalSupply(c context.Context, _ *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	totalSupply := q.GetSupply(ctx).GetTotal()

	return &types.QueryTotalSupplyResponse{Supply: totalSupply}, nil
}

// SupplyOf implements the Query/SupplyOf gRPC method
func (q BaseKeeper) SupplyOf(c context.Context, req *types.QuerySupplyOfRequest) (*types.QuerySupplyOfResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Denom == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(c)
	supply := q.GetSupply(ctx).GetTotal().AmountOf(req.Denom)

	return &types.QuerySupplyOfResponse{Amount: supply}, nil
}
