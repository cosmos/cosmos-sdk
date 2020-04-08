package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

func (q Querier) QueryBalance(ctx context.Context, params *types.QueryBalanceParams) (*sdk.Coin, error) {
	balance := q.GetBalance(sdk.UnwrapSDKContext(ctx), params.Address, params.Denom)
	return &balance, nil
}

func (q Querier) QueryAllBalances(ctx context.Context, params *types.QueryAllBalancesParams) (*types.QueryAllBalancesResponse, error) {
	balances := q.GetAllBalances(sdk.UnwrapSDKContext(ctx), params.Address)
	return &types.QueryAllBalancesResponse{Balances: balances}, nil
}
