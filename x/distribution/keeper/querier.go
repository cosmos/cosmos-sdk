package keeper

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, path[1:], req, k)

		case types.QueryValidatorOutstandingRewards:
			return queryValidatorOutstandingRewards(ctx, path[1:], req, k)

		case types.QueryValidatorCommission:
			return queryValidatorCommission(ctx, path[1:], req, k)

		case types.QueryValidatorSlashes:
			return queryValidatorSlashes(ctx, path[1:], req, k)

		case types.QueryDelegationRewards:
			return queryDelegationRewards(ctx, path[1:], req, k)

		case types.QueryDelegatorTotalRewards:
			return queryDelegatorTotalRewards(ctx, path[1:], req, k)

		case types.QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, path[1:], req, k)

		case types.QueryWithdrawAddr:
			return queryDelegatorWithdrawAddress(ctx, path[1:], req, k)

		case types.QueryCommunityPool:
			return queryCommunityPool(ctx, path[1:], req, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(k.cdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryValidatorOutstandingRewards(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorOutstandingRewardsParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	rewards := k.GetValidatorOutstandingRewards(ctx, params.ValidatorAddress)
	if rewards == nil {
		rewards = sdk.DecCoins{}
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, rewards)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryValidatorCommission(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorCommissionParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	commission := k.GetValidatorAccumulatedCommission(ctx, params.ValidatorAddress)
	if commission == nil {
		commission = sdk.DecCoins{}
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, commission)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryValidatorSlashes(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorSlashesParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	events := make([]types.ValidatorSlashEvent, 0)
	k.IterateValidatorSlashEventsBetween(ctx, params.ValidatorAddress, params.StartingHeight, params.EndingHeight,
		func(height uint64, event types.ValidatorSlashEvent) (stop bool) {
			events = append(events, event)
			return false
		},
	)

	bz, err := codec.MarshalJSONIndent(k.cdc, events)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegationRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegationRewardsParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	val := k.stakingKeeper.Validator(ctx, params.ValidatorAddress)
	if val == nil {
		return nil, sdkerrors.Wrap(types.ErrNoValidatorExists, params.ValidatorAddress.String())
	}

	del := k.stakingKeeper.Delegation(ctx, params.DelegatorAddress, params.ValidatorAddress)
	if del == nil {
		return nil, types.ErrNoDelegationExists
	}

	endingPeriod := k.incrementValidatorPeriod(ctx, val)
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)
	if rewards == nil {
		rewards = sdk.DecCoins{}
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, rewards)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegatorTotalRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	total := sdk.DecCoins{}
	var delRewards []types.DelegationDelegatorReward

	k.stakingKeeper.IterateDelegations(
		ctx, params.DelegatorAddress,
		func(_ int64, del exported.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(ctx, valAddr)
			endingPeriod := k.incrementValidatorPeriod(ctx, val)
			delReward := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

			delRewards = append(delRewards, types.NewDelegationDelegatorReward(valAddr, delReward))
			total = total.Add(delReward...)
			return false
		},
	)

	totalRewards := types.NewQueryDelegatorTotalRewardsResponse(delRewards, total)

	bz, err := json.Marshal(totalRewards)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegatorValidators(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	var validators []sdk.ValAddress

	k.stakingKeeper.IterateDelegations(
		ctx, params.DelegatorAddress,
		func(_ int64, del exported.DelegationI) (stop bool) {
			validators = append(validators, del.GetValidatorAddr())
			return false
		},
	)

	bz, err := codec.MarshalJSONIndent(k.cdc, validators)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegatorWithdrawAddress(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorWithdrawAddrParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, params.DelegatorAddress)

	bz, err := codec.MarshalJSONIndent(k.cdc, withdrawAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryCommunityPool(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	pool := k.GetFeePoolCommunityCoins(ctx)
	if pool == nil {
		pool = sdk.DecCoins{}
	}

	bz, err := k.cdc.MarshalJSON(pool)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
