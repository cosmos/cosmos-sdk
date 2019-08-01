package keeper

import (
	"fmt"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// creates a querier for staking REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryValidators:
			return queryValidators(ctx, req, k)
		case types.QueryValidator:
			return queryValidator(ctx, req, k)
		case types.QueryValidatorDelegations:
			return queryValidatorDelegations(ctx, req, k)
		case types.QueryValidatorUnbondingDelegations:
			return queryValidatorUnbondingDelegations(ctx, req, k)
		case types.QueryDelegation:
			return queryDelegation(ctx, req, k)
		case types.QueryUnbondingDelegation:
			return queryUnbondingDelegation(ctx, req, k)
		case types.QueryDelegatorDelegations:
			return queryDelegatorDelegations(ctx, req, k)
		case types.QueryDelegatorUnbondingDelegations:
			return queryDelegatorUnbondingDelegations(ctx, req, k)
		case types.QueryRedelegations:
			return queryRedelegations(ctx, req, k)
		case types.QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, req, k)
		case types.QueryDelegatorValidator:
			return queryDelegatorValidator(ctx, req, k)
		case types.QueryPool:
			return queryPool(ctx, k)
		case types.QueryParameters:
			return queryParameters(ctx, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown staking query endpoint")
		}
	}
}

func queryValidators(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryValidatorsParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	validators := k.GetAllValidators(ctx)
	filteredVals := make([]types.Validator, 0, len(validators))

	for _, val := range validators {
		if strings.ToLower(val.GetStatus().String()) == strings.ToLower(params.Status) {
			filteredVals = append(filteredVals, val)
		}
	}

	start, end := client.Paginate(len(filteredVals), params.Page, params.Limit, int(k.GetParams(ctx).MaxValidators))
	if start < 0 || end < 0 {
		filteredVals = []types.Validator{}
	} else {
		filteredVals = filteredVals[start:end]
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, filteredVals)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

func queryValidator(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryValidatorParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	validator, found := k.GetValidator(ctx, params.ValidatorAddr)
	if !found {
		return nil, types.ErrNoValidatorFound(types.DefaultCodespace)
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, validator)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryValidatorDelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryValidatorParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	delegations := k.GetValidatorDelegations(ctx, params.ValidatorAddr)
	delegationResps, err := delegationsToDelegationResponses(ctx, k, delegations)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, delegationResps)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryValidatorUnbondingDelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryValidatorParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	unbonds := k.GetUnbondingDelegationsFromValidator(ctx, params.ValidatorAddr)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, unbonds)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryDelegatorDelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryDelegatorParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	delegations := k.GetAllDelegatorDelegations(ctx, params.DelegatorAddr)
	delegationResps, err := delegationsToDelegationResponses(ctx, k, delegations)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, delegationResps)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryDelegatorUnbondingDelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryDelegatorParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	unbondingDelegations := k.GetAllUnbondingDelegations(ctx, params.DelegatorAddr)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, unbondingDelegations)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryDelegatorValidators(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryDelegatorParams

	stakingParams := k.GetParams(ctx)

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	validators := k.GetDelegatorValidators(ctx, params.DelegatorAddr, stakingParams.MaxValidators)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, validators)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryDelegatorValidator(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryBondsParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	validator, err := k.GetDelegatorValidator(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, validator)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryDelegation(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryBondsParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	delegation, found := k.GetDelegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if !found {
		return nil, types.ErrNoDelegation(types.DefaultCodespace)
	}

	delegationResp, err := delegationToDelegationResponse(ctx, k, delegation)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, delegationResp)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryUnbondingDelegation(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryBondsParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	unbond, found := k.GetUnbondingDelegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation(types.DefaultCodespace)
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, unbond)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryRedelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryRedelegationParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(string(req.Data))
	}

	var redels []types.Redelegation

	if !params.DelegatorAddr.Empty() && !params.SrcValidatorAddr.Empty() && !params.DstValidatorAddr.Empty() {
		redel, found := k.GetRedelegation(ctx, params.DelegatorAddr, params.SrcValidatorAddr, params.DstValidatorAddr)
		if !found {
			return nil, types.ErrNoRedelegation(types.DefaultCodespace)
		}

		redels = []types.Redelegation{redel}
	} else if params.DelegatorAddr.Empty() && !params.SrcValidatorAddr.Empty() && params.DstValidatorAddr.Empty() {
		redels = k.GetRedelegationsFromSrcValidator(ctx, params.SrcValidatorAddr)
	} else {
		redels = k.GetAllRedelegations(ctx, params.DelegatorAddr, params.SrcValidatorAddr, params.DstValidatorAddr)
	}

	redelResponses, err := redelegationsToRedelegationResponses(ctx, k, redels)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, redelResponses)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryPool(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	bondDenom := k.BondDenom(ctx)
	bondedPool := k.GetBondedPool(ctx)
	notBondedPool := k.GetNotBondedPool(ctx)
	if bondedPool == nil || notBondedPool == nil {
		return nil, sdk.ErrInternal("pool accounts haven't been set")
	}

	pool := types.NewPool(
		notBondedPool.GetCoins().AmountOf(bondDenom),
		bondedPool.GetCoins().AmountOf(bondDenom),
	)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, pool)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryParameters(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

//______________________________________________________
// util

func delegationToDelegationResponse(ctx sdk.Context, k Keeper, del types.Delegation) (types.DelegationResponse, sdk.Error) {
	val, found := k.GetValidator(ctx, del.ValidatorAddress)
	if !found {
		return types.DelegationResponse{}, types.ErrNoValidatorFound(types.DefaultCodespace)
	}

	return types.NewDelegationResp(
		del.DelegatorAddress,
		del.ValidatorAddress,
		del.Shares,
		val.TokensFromShares(del.Shares).TruncateInt(),
	), nil
}

func delegationsToDelegationResponses(
	ctx sdk.Context, k Keeper, delegations types.Delegations,
) (types.DelegationResponses, sdk.Error) {

	resp := make(types.DelegationResponses, len(delegations), len(delegations))
	for i, del := range delegations {
		delResp, err := delegationToDelegationResponse(ctx, k, del)
		if err != nil {
			return nil, err
		}

		resp[i] = delResp
	}

	return resp, nil
}

func redelegationsToRedelegationResponses(
	ctx sdk.Context, k Keeper, redels types.Redelegations,
) (types.RedelegationResponses, sdk.Error) {

	resp := make(types.RedelegationResponses, len(redels), len(redels))
	for i, redel := range redels {
		val, found := k.GetValidator(ctx, redel.ValidatorDstAddress)
		if !found {
			return nil, types.ErrNoValidatorFound(types.DefaultCodespace)
		}

		entryResponses := make([]types.RedelegationEntryResponse, len(redel.Entries), len(redel.Entries))
		for j, entry := range redel.Entries {
			entryResponses[j] = types.NewRedelegationEntryResponse(
				entry.CreationHeight,
				entry.CompletionTime,
				entry.SharesDst,
				entry.InitialBalance,
				val.TokensFromShares(entry.SharesDst).TruncateInt(),
			)
		}

		resp[i] = types.NewRedelegationResponse(
			redel.DelegatorAddress,
			redel.ValidatorSrcAddress,
			redel.ValidatorDstAddress,
			entryResponses,
		)
	}

	return resp, nil
}
