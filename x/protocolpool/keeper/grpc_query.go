package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

// CommunityPool queries the community pool coins
func (k Querier) CommunityPool(ctx context.Context, _ *types.QueryCommunityPoolRequest) (*types.QueryCommunityPoolResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	amount, err := k.Keeper.GetCommunityPool(sdkCtx)
	if err != nil {
		return nil, err
	}
	decCoins := sdk.NewDecCoinsFromCoins(amount...)
	return &types.QueryCommunityPoolResponse{Pool: decCoins}, nil
}

// UnclaimedBudget queries the unclaimed budget for given recipient
func (k Querier) UnclaimedBudget(ctx context.Context, req *types.QueryUnclaimedBudgetRequest) (*types.QueryUnclaimedBudgetResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	address, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid recipient address: %s", err.Error())
	}
	budget, err := k.Keeper.Budgets.Get(sdkCtx, address)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "no budget proposal found for address: %s", req.Address)
		}
		return nil, err
	}

	totalBudgetAmountLeftToDistribute := budget.BudgetPerTranche.Amount.Mul(math.NewIntFromUint64(budget.TranchesLeft))
	totalBudgetAmountLeft := sdk.NewCoin(budget.BudgetPerTranche.Denom, totalBudgetAmountLeftToDistribute)

	var unclaimedBudget sdk.Coin
	if budget.ClaimedAmount == nil {
		unclaimedBudget = totalBudgetAmountLeft
		zeroCoin := sdk.NewCoin(budget.BudgetPerTranche.Denom, math.ZeroInt())
		budget.ClaimedAmount = &zeroCoin
	} else {
		unclaimedBudget = totalBudgetAmountLeft
	}

	nextClaimFrom := budget.LastClaimedAt.Add(*budget.Period)

	return &types.QueryUnclaimedBudgetResponse{
		ClaimedAmount:   budget.ClaimedAmount,
		UnclaimedAmount: &unclaimedBudget,
		NextClaimFrom:   &nextClaimFrom,
		Period:          budget.Period,
		TranchesLeft:    budget.TranchesLeft,
	}, nil
}

// Params queries params of x/protocolpool module.
func (k Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	params, err := k.Keeper.Params.Get(sdkCtx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "params not found")
		}
		return nil, err
	}

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
