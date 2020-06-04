package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ types.QueryServer = BaseKeeper{}

// Balance implements the Query/Balance gRPC method
func (q BaseKeeper) Balance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	balance := q.GetBalance(sdk.UnwrapSDKContext(ctx), req.Address, req.Denom)
	return &types.QueryBalanceResponse{Balance: &balance}, nil
}

// AllBalances implements the Query/AllBalances gRPC method
func (q BaseKeeper) AllBalances(c context.Context, req *types.QueryAllBalancesRequest) (*types.QueryAllBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	balances := q.GetAllBalances(ctx, req.Address)

	return &types.QueryAllBalancesResponse{Balances: balances}, nil
}

// TotalSupply implements the Query/TotalSupply gRPC method
func (q BaseKeeper) TotalSupply(c context.Context, request *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// TODO: add pagination once it can be supported properly with supply no longer being stored as a single blob
	totalSupply := q.GetSupply(ctx).GetTotal()

	return &types.QueryTotalSupplyResponse{Supply: totalSupply}, nil
}

// SupplyOf implements the Query/SupplyOf gRPC method
func (q BaseKeeper) SupplyOf(c context.Context, request *types.QuerySupplyOfRequest) (*types.QuerySupplyOfResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	supply := q.GetSupply(ctx).GetTotal().AmountOf(request.Denom)

	return &types.QuerySupplyOfResponse{Amount: supply}, nil
}
