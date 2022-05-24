package keeper

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, path[1:], req, k, legacyQuerierCdc)

		case types.QueryValidatorOutstandingRewards:
			return queryValidatorOutstandingRewards(ctx, path[1:], req, k, legacyQuerierCdc)

		case types.QueryValidatorCommission:
			return queryValidatorCommission(ctx, path[1:], req, k, legacyQuerierCdc)

		case types.QueryValidatorSlashes:
			return queryValidatorSlashes(ctx, path[1:], req, k, legacyQuerierCdc)

		case types.QueryDelegationRewards:
			return queryDelegationRewards(ctx, path[1:], req, k, legacyQuerierCdc)

		case types.QueryDelegatorTotalRewards:
			return queryDelegatorTotalRewards(ctx, path[1:], req, k, legacyQuerierCdc)

		case types.QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, path[1:], req, k, legacyQuerierCdc)

		case types.QueryWithdrawAddr:
			return queryDelegatorWithdrawAddress(ctx, path[1:], req, k, legacyQuerierCdc)

		case types.QueryCommunityPool:
			return queryCommunityPool(ctx, path[1:], req, k, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryParams(ctx sdk.Context, _ []string, _ abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryValidatorOutstandingRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryValidatorOutstandingRewardsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	rewards := k.GetValidatorOutstandingRewards(ctx, params.ValidatorAddress)
	if rewards.GetRewards() == nil {
		rewards.Rewards = sdk.DecCoins{}
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, rewards)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryValidatorCommission(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryValidatorCommissionParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	commission := k.GetValidatorAccumulatedCommission(ctx, params.ValidatorAddress)
	if commission.Commission == nil {
		commission.Commission = sdk.DecCoins{}
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, commission)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryValidatorSlashes(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryValidatorSlashesParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
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

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, events)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegationRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegationRewardsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// branch the context to isolate state changes
	ctx, _ = ctx.CacheContext()

	val := k.stakingKeeper.Validator(ctx, params.ValidatorAddress)
	if val == nil {
		return nil, sdkerrors.Wrap(types.ErrNoValidatorExists, params.ValidatorAddress.String())
	}

	del := k.stakingKeeper.Delegation(ctx, params.DelegatorAddress, params.ValidatorAddress)
	if del == nil {
		return nil, types.ErrNoDelegationExists
	}

	endingPeriod := k.IncrementValidatorPeriod(ctx, val)
	rewards := k.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	if rewards == nil {
		rewards = sdk.DecCoins{}
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, rewards)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegatorTotalRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// branch the context to isolate state changes
	ctx, _ = ctx.CacheContext()

	total := sdk.DecCoins{}

	var delRewards []types.DelegationDelegatorReward

	k.stakingKeeper.IterateDelegations(
		ctx, params.DelegatorAddress,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(ctx, valAddr)
			endingPeriod := k.IncrementValidatorPeriod(ctx, val)
			delReward := k.CalculateDelegationRewards(ctx, val, del, endingPeriod)

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

func queryDelegatorValidators(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// branch the context to isolate state changes
	ctx, _ = ctx.CacheContext()

	var validators []sdk.ValAddress

	k.stakingKeeper.IterateDelegations(
		ctx, params.DelegatorAddress,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			validators = append(validators, del.GetValidatorAddr())
			return false
		},
	)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, validators)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegatorWithdrawAddress(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorWithdrawAddrParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// branch the context to isolate state changes
	ctx, _ = ctx.CacheContext()
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, params.DelegatorAddress)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, withdrawAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryCommunityPool(ctx sdk.Context, _ []string, _ abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	pool := k.GetFeePoolCommunityCoins(ctx)
	if pool == nil {
		pool = sdk.DecCoins{}
	}

	bz, err := legacyQuerierCdc.MarshalJSON(pool)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
