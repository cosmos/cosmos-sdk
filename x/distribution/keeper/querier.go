package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// nolint
const (
	QueryParams                = "params"
	QueryOutstandingRewards    = "outstanding_rewards"
	QueryValidatorCommission   = "validator_commission"
	QueryValidatorSlashes      = "validator_slashes"
	QueryDelegationRewards     = "delegation_rewards"
	QueryDelegatorTotalRewards = "delegator_total_rewards"
	QueryDelegatorValidators   = "delegator_validators"
	QueryWithdrawAddr          = "withdraw_addr"

	ParamCommunityTax        = "community_tax"
	ParamBaseProposerReward  = "base_proposer_reward"
	ParamBonusProposerReward = "bonus_proposer_reward"
	ParamWithdrawAddrEnabled = "withdraw_addr_enabled"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryParams:
			return queryParams(ctx, path[1:], req, k)

		case QueryOutstandingRewards:
			return queryOutstandingRewards(ctx, path[1:], req, k)

		case QueryValidatorCommission:
			return queryValidatorCommission(ctx, path[1:], req, k)

		case QueryValidatorSlashes:
			return queryValidatorSlashes(ctx, path[1:], req, k)

		case QueryDelegationRewards:
			return queryDelegationRewards(ctx, path[1:], req, k)

		case QueryDelegatorTotalRewards:
			return queryDelegatorTotalRewards(ctx, path[1:], req, k)

		case QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, path[1:], req, k)

		case QueryWithdrawAddr:
			return queryDelegatorWithdrawAddress(ctx, path[1:], req, k)

		default:
			return nil, sdk.ErrUnknownRequest("unknown distr query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	switch path[0] {
	case ParamCommunityTax:
		bz, err := codec.MarshalJSONIndent(k.cdc, k.GetCommunityTax(ctx))
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	case ParamBaseProposerReward:
		bz, err := codec.MarshalJSONIndent(k.cdc, k.GetBaseProposerReward(ctx))
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	case ParamBonusProposerReward:
		bz, err := codec.MarshalJSONIndent(k.cdc, k.GetBonusProposerReward(ctx))
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	case ParamWithdrawAddrEnabled:
		bz, err := codec.MarshalJSONIndent(k.cdc, k.GetWithdrawAddrEnabled(ctx))
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	default:
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}
}

func queryOutstandingRewards(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	bz, err := codec.MarshalJSONIndent(k.cdc, k.GetOutstandingRewards(ctx))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// params for query 'custom/distr/validator_commission'
type QueryValidatorCommissionParams struct {
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
}

// creates a new instance of QueryValidatorCommissionParams
func NewQueryValidatorCommissionParams(validatorAddr sdk.ValAddress) QueryValidatorCommissionParams {
	return QueryValidatorCommissionParams{
		ValidatorAddr: validatorAddr,
	}
}

func queryValidatorCommission(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryValidatorCommissionParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	commission := k.GetValidatorAccumulatedCommission(ctx, params.ValidatorAddr)
	bz, err := codec.MarshalJSONIndent(k.cdc, commission)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// params for query 'custom/distr/validator_slashes'
type QueryValidatorSlashesParams struct {
	ValidatorAddr  sdk.ValAddress `json:"validator_addr"`
	StartingHeight uint64         `json:"starting_height"`
	EndingHeight   uint64         `json:"ending_height"`
}

// creates a new instance of QueryValidatorSlashesParams
func NewQueryValidatorSlashesParams(validatorAddr sdk.ValAddress, startingHeight uint64, endingHeight uint64) QueryValidatorSlashesParams {
	return QueryValidatorSlashesParams{
		ValidatorAddr:  validatorAddr,
		StartingHeight: startingHeight,
		EndingHeight:   endingHeight,
	}
}

func queryValidatorSlashes(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryValidatorSlashesParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	events := make([]types.ValidatorSlashEvent, 0)
	k.IterateValidatorSlashEventsBetween(ctx, params.ValidatorAddr, params.StartingHeight, params.EndingHeight,
		func(height uint64, event types.ValidatorSlashEvent) (stop bool) {
			events = append(events, event)
			return false
		},
	)
	bz, err := codec.MarshalJSONIndent(k.cdc, events)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// params for query 'custom/distr/delegation_rewards'
type QueryDelegationRewardsParams struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
}

// creates a new instance of QueryDelegationRewardsParams
func NewQueryDelegationRewardsParams(delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) QueryDelegationRewardsParams {
	return QueryDelegationRewardsParams{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
	}
}

func queryDelegationRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryDelegationRewardsParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	val := k.stakingKeeper.Validator(ctx, params.ValidatorAddr)
	del := k.stakingKeeper.Delegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	endingPeriod := k.incrementValidatorPeriod(ctx, val)
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	bz, err := codec.MarshalJSONIndent(k.cdc, rewards)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// params for query 'custom/distr/delegator_total_rewards' and 'custom/distr/delegator_validators'
type QueryDelegatorParams struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
}

// creates a new instance of QueryDelegationRewardsParams
func NewQueryDelegatorParams(delegatorAddr sdk.AccAddress) QueryDelegatorParams {
	return QueryDelegatorParams{
		DelegatorAddr: delegatorAddr,
	}
}

func queryDelegatorTotalRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryDelegatorParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	var totalRewards sdk.DecCoins

	k.stakingKeeper.IterateDelegations(
		ctx, params.DelegatorAddr,
		func(_ int64, del sdk.Delegation) (stop bool) {
			val := k.stakingKeeper.Validator(ctx, del.GetValidatorAddr())
			endingPeriod := k.incrementValidatorPeriod(ctx, val)
			rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

			totalRewards = totalRewards.Plus(rewards)
			return false
		},
	)

	bz, err := codec.MarshalJSONIndent(k.cdc, totalRewards)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryDelegatorValidators(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryDelegatorParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	var validators []sdk.ValAddress

	k.stakingKeeper.IterateDelegations(
		ctx, params.DelegatorAddr,
		func(_ int64, del sdk.Delegation) (stop bool) {
			validators = append(validators[:], del.GetValidatorAddr())
			return false
		},
	)

	bz, err := codec.MarshalJSONIndent(k.cdc, validators)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// params for query 'custom/distr/withdraw_addr'
type QueryDelegatorWithdrawAddrParams struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
}

// NewQueryDelegatorWithdrawAddrParams creates a new instance of QueryDelegatorWithdrawAddrParams.
func NewQueryDelegatorWithdrawAddrParams(delegatorAddr sdk.AccAddress) QueryDelegatorWithdrawAddrParams {
	return QueryDelegatorWithdrawAddrParams{DelegatorAddr: delegatorAddr}
}

func queryDelegatorWithdrawAddress(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryDelegatorWithdrawAddrParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, params.DelegatorAddr)

	bz, err := codec.MarshalJSONIndent(k.cdc, withdrawAddr)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
