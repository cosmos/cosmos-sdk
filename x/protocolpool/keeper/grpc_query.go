package keeper

import (
	"context"

	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

// CommunityPool queries the community pool coins
func (k Querier) CommunityPool(ctx context.Context, req *types.QueryCommunityPoolRequest) (*types.QueryCommunityPoolResponse, error) {
	amount := k.Keeper.GetCommunityPool(ctx)
	decCoins := sdk.NewDecCoinsFromCoins(amount...)
	return &types.QueryCommunityPoolResponse{Pool: decCoins}, nil
}
