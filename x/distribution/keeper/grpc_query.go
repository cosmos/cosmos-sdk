package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// Params queries params of distribution module
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ValidatorOutstandingRewards queries rewards of a validator address
func (k Keeper) ValidatorOutstandingRewards(c context.Context, req *types.QueryValidatorOutstandingRewardsRequest) (*types.QueryValidatorOutstandingRewardsResponse, error) {
	if req.String() == "" || req.ValidatorAddress.Empty() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	rewards := k.GetValidatorOutstandingRewards(ctx, req.ValidatorAddress)

	return &types.QueryValidatorOutstandingRewardsResponse{Rewards: rewards}, nil
}

// ValidatorCommission queries accumulated commission for a validator
func (k Keeper) ValidatorCommission(c context.Context, req *types.QueryValidatorCommissionRequest) (*types.QueryValidatorCommissionResponse, error) {
	if req.String() == "" || req.ValidatorAddress.Empty() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	commission := k.GetValidatorAccumulatedCommission(ctx, req.ValidatorAddress)

	return &types.QueryValidatorCommissionResponse{Commission: commission}, nil
}

// ValidatorSlashes queries slash events of a validator
func (k Keeper) ValidatorSlashes(c context.Context, req *types.QueryValidatorSlashesRequest) (*types.QueryValidatorSlashesResponse, error) {
	if req.String() == "" || req.ValidatorAddress.Empty() || req.EndingHeight < req.StartingHeight {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	events := make([]types.ValidatorSlashEvent, 0)
	store := ctx.KVStore(k.storeKey)
	slashesStore := prefix.NewStore(store, types.GetValidatorSlashEventPrefix(req.ValidatorAddress))

	res, err := query.FilteredPaginate(slashesStore, req.Req, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var result types.ValidatorSlashEvent
		err := k.cdc.UnmarshalBinaryBare(value, &result)

		if err != nil {
			return false, err
		}

		if result.ValidatorPeriod < req.StartingHeight || result.ValidatorPeriod > req.EndingHeight {
			return false, nil
		}

		if accumulate {
			events = append(events, result)
		}
		return true, nil
	})

	if err != nil {
		return &types.QueryValidatorSlashesResponse{}, err
	}

	return &types.QueryValidatorSlashesResponse{Slashes: events, Res: res}, nil
}

// DelegationRewards the total rewards accrued by a delegation
func (k Keeper) DelegationRewards(c context.Context, req *types.QueryDelegationRewardsRequest) (*types.QueryDelegationRewardsResponse, error) {
	if req.String() == "" || req.DelegatorAddress.Empty() || req.ValidatorAddress.Empty() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	val := k.stakingKeeper.Validator(ctx, req.ValidatorAddress)
	if val == nil {
		return nil, sdkerrors.Wrap(types.ErrNoValidatorExists, req.ValidatorAddress.String())
	}

	del := k.stakingKeeper.Delegation(ctx, req.DelegatorAddress, req.ValidatorAddress)
	if del == nil {
		return nil, types.ErrNoDelegationExists
	}

	endingPeriod := k.IncrementValidatorPeriod(ctx, val)
	rewards := k.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	if rewards == nil {
		rewards = sdk.DecCoins{}
	}

	return &types.QueryDelegationRewardsResponse{Rewards: rewards}, nil
}

// DelegationTotalRewards the total rewards accrued by a each validator
func (k Keeper) DelegationTotalRewards(c context.Context, req *types.QueryDelegationTotalRewardsRequest) (*types.QueryDelegationTotalRewardsResponse, error) {
	if req.String() == "" || req.DelegatorAddress.Empty() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	total := sdk.DecCoins{}
	var delRewards []types.DelegationDelegatorReward

	k.stakingKeeper.IterateDelegations(
		ctx, req.DelegatorAddress,
		func(_ int64, del exported.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(ctx, valAddr)
			endingPeriod := k.IncrementValidatorPeriod(ctx, val)
			delReward := k.CalculateDelegationRewards(ctx, val, del, endingPeriod)

			delRewards = append(delRewards, types.NewDelegationDelegatorReward(valAddr, delReward))
			total = total.Add(delReward...)
			return false
		},
	)

	return &types.QueryDelegationTotalRewardsResponse{Rewards: delRewards, Total: total}, nil
}

// DelegatorValidators queries the validators list of a delegator
func (k Keeper) DelegatorValidators(c context.Context, req *types.QueryDelegatorValidatorsRequest) (*types.QueryDelegatorValidatorsResponse, error) {
	if req.String() == "" || req.DelegatorAddress.Empty() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var validators []sdk.ValAddress

	k.stakingKeeper.IterateDelegations(
		ctx, req.DelegatorAddress,
		func(_ int64, del exported.DelegationI) (stop bool) {
			validators = append(validators, del.GetValidatorAddr())
			return false
		},
	)

	return &types.QueryDelegatorValidatorsResponse{Validators: validators}, nil
}

// DelegatorWithdrawAddress queries Query/delegatorWithdrawAddress
func (k Keeper) DelegatorWithdrawAddress(c context.Context, req *types.QueryDelegatorWithdrawAddressRequest) (*types.QueryDelegatorWithdrawAddressResponse, error) {
	if req.String() == "" || req.DelegatorAddress.Empty() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, req.DelegatorAddress)

	return &types.QueryDelegatorWithdrawAddressResponse{WithdrawAddress: withdrawAddr}, nil
}

// CommunityPool queries the community pool coins
func (k Keeper) CommunityPool(c context.Context, req *types.QueryCommunityPoolRequest) (*types.QueryCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pool := k.GetFeePoolCommunityCoins(ctx)
	if pool == nil {
		pool = sdk.DecCoins{}
	}

	return &types.QueryCommunityPoolResponse{Pool: pool}, nil
}
