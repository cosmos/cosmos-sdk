package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type Querier struct {
	Keeper
}

var _ types.QueryServiceServer = Querier{}

func (q Querier) QueryBalance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	balance := q.GetBalance(sdk.UnwrapSDKContext(ctx), req.Address, req.Denom)
	return &types.QueryBalanceResponse{Balance: &balance}, nil
}

func (q Querier) QueryAllBalances(ctx context.Context, req *types.QueryAllBalancesRequest) (*types.QueryAllBalancesResponse, error) {
	balances := q.GetAllBalances(sdk.UnwrapSDKContext(ctx), req.Address)
	return &types.QueryAllBalancesResponse{Balances: balances}, nil
}
