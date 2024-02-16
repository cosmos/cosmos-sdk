package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
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
	amount, err := k.Keeper.GetCommunityPool(ctx)
	if err != nil {
		return nil, err
	}
	decCoins := sdk.NewDecCoinsFromCoins(amount...)
	return &types.QueryCommunityPoolResponse{Pool: decCoins}, nil
}

// UnclaimedBudget queries the unclaimed budget for given recipient
func (k Querier) UnclaimedBudget(ctx context.Context, req *types.QueryUnclaimedBudgetRequest) (*types.QueryUnclaimedBudgetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	address, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid recipient address: %s", err.Error())
	}
	budget, err := k.Keeper.BudgetProposal.Get(ctx, address)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "no budget proposal found for address: %s", req.Address)
		}
		return nil, err
	}
	var unclaimedBudget sdk.Coin
	if budget.ClaimedAmount == nil {
		unclaimedBudget = *budget.TotalBudget
		zeroCoin := sdk.NewCoin(budget.TotalBudget.Denom, math.ZeroInt())
		budget.ClaimedAmount = &zeroCoin
	} else {
		unclaimedBudget = budget.TotalBudget.Sub(*budget.ClaimedAmount)
	}

	if budget.NextClaimFrom == nil {
		budget.NextClaimFrom = budget.StartTime
	}

	if budget.TranchesLeft == 0 {
		budget.TranchesLeft = budget.Tranches
	}

	return &types.QueryUnclaimedBudgetResponse{
		TotalBudget:     budget.TotalBudget,
		ClaimedAmount:   budget.ClaimedAmount,
		UnclaimedAmount: &unclaimedBudget,
		NextClaimFrom:   budget.NextClaimFrom,
		Period:          budget.Period,
		TranchesLeft:    budget.TranchesLeft,
	}, nil
}
