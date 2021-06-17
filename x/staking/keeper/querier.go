package keeper

import (
	"errors"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// creates a querier for staking REST endpoints
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryValidators:
			return queryValidators(ctx, req, k, legacyQuerierCdc)

		case types.QueryValidator:
			return queryValidator(ctx, req, k, legacyQuerierCdc)

		case types.QueryValidatorDelegations:
			return queryValidatorDelegations(ctx, req, k, legacyQuerierCdc)

		case types.QueryValidatorUnbondingDelegations:
			return queryValidatorUnbondingDelegations(ctx, req, k, legacyQuerierCdc)

		case types.QueryDelegation:
			return queryDelegation(ctx, req, k, legacyQuerierCdc)

		case types.QueryUnbondingDelegation:
			return queryUnbondingDelegation(ctx, req, k, legacyQuerierCdc)

		case types.QueryDelegatorDelegations:
			return queryDelegatorDelegations(ctx, req, k, legacyQuerierCdc)

		case types.QueryDelegatorUnbondingDelegations:
			return queryDelegatorUnbondingDelegations(ctx, req, k, legacyQuerierCdc)

		case types.QueryRedelegations:
			return queryRedelegations(ctx, req, k, legacyQuerierCdc)

		case types.QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, req, k, legacyQuerierCdc)

		case types.QueryDelegatorValidator:
			return queryDelegatorValidator(ctx, req, k, legacyQuerierCdc)

		case types.QueryHistoricalInfo:
			return queryHistoricalInfo(ctx, req, k, legacyQuerierCdc)

		case types.QueryPool:
			return queryPool(ctx, k, legacyQuerierCdc)

		case types.QueryParameters:
			return queryParameters(ctx, k, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

func queryValidators(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryValidatorsParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	validators := k.GetAllValidators(ctx)
	filteredVals := make(types.Validators, 0, len(validators))

	for _, val := range validators {
		if strings.EqualFold(val.GetStatus().String(), params.Status) {
			filteredVals = append(filteredVals, val)
		}
	}

	start, end := client.Paginate(len(filteredVals), params.Page, params.Limit, int(k.GetParams(ctx).MaxValidators))
	if start < 0 || end < 0 {
		filteredVals = []types.Validator{}
	} else {
		filteredVals = filteredVals[start:end]
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, filteredVals)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryValidator(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryValidatorParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	validator, found := k.GetValidator(ctx, params.ValidatorAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, validator)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryValidatorDelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryValidatorParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	delegations := k.GetValidatorDelegations(ctx, params.ValidatorAddr)

	start, end := client.Paginate(len(delegations), params.Page, params.Limit, int(k.GetParams(ctx).MaxValidators))
	if start < 0 || end < 0 {
		delegations = []types.Delegation{}
	} else {
		delegations = delegations[start:end]
	}

	delegationResps, err := DelegationsToDelegationResponses(ctx, k, delegations)
	if err != nil {
		return nil, err
	}

	if delegationResps == nil {
		delegationResps = types.DelegationResponses{}
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, delegationResps)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryValidatorUnbondingDelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryValidatorParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	unbonds := k.GetUnbondingDelegationsFromValidator(ctx, params.ValidatorAddr)
	if unbonds == nil {
		unbonds = types.UnbondingDelegations{}
	}

	start, end := client.Paginate(len(unbonds), params.Page, params.Limit, int(k.GetParams(ctx).MaxValidators))
	if start < 0 || end < 0 {
		unbonds = types.UnbondingDelegations{}
	} else {
		unbonds = unbonds[start:end]
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, unbonds)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryDelegatorDelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	delegations := k.GetAllDelegatorDelegations(ctx, params.DelegatorAddr)
	delegationResps, err := DelegationsToDelegationResponses(ctx, k, delegations)

	if err != nil {
		return nil, err
	}

	if delegationResps == nil {
		delegationResps = types.DelegationResponses{}
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, delegationResps)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryDelegatorUnbondingDelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	unbondingDelegations := k.GetAllUnbondingDelegations(ctx, params.DelegatorAddr)
	if unbondingDelegations == nil {
		unbondingDelegations = types.UnbondingDelegations{}
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, unbondingDelegations)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryDelegatorValidators(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorParams

	stakingParams := k.GetParams(ctx)

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	validators := k.GetDelegatorValidators(ctx, params.DelegatorAddr, stakingParams.MaxValidators)
	if validators == nil {
		validators = types.Validators{}
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, validators)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryDelegatorValidator(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorValidatorRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	delAddr, err := sdk.AccAddressFromBech32(params.DelegatorAddr)
	if err != nil {
		return nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(params.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	validator, err := k.GetDelegatorValidator(ctx, delAddr, valAddr)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, validator)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryDelegation(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorValidatorRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	delAddr, err := sdk.AccAddressFromBech32(params.DelegatorAddr)
	if err != nil {
		return nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(params.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	delegation, found := k.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil, types.ErrNoDelegation
	}

	delegationResp, err := DelegationToDelegationResponse(ctx, k, delegation)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, delegationResp)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryUnbondingDelegation(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDelegatorValidatorRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	delAddr, err := sdk.AccAddressFromBech32(params.DelegatorAddr)
	if err != nil {
		return nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(params.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	unbond, found := k.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, unbond)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryRedelegations(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryRedelegationParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	var redels []types.Redelegation

	switch {
	case !params.DelegatorAddr.Empty() && !params.SrcValidatorAddr.Empty() && !params.DstValidatorAddr.Empty():
		redel, found := k.GetRedelegation(ctx, params.DelegatorAddr, params.SrcValidatorAddr, params.DstValidatorAddr)
		if !found {
			return nil, types.ErrNoRedelegation
		}

		redels = []types.Redelegation{redel}
	case params.DelegatorAddr.Empty() && !params.SrcValidatorAddr.Empty() && params.DstValidatorAddr.Empty():
		redels = k.GetRedelegationsFromSrcValidator(ctx, params.SrcValidatorAddr)
	default:
		redels = k.GetAllRedelegations(ctx, params.DelegatorAddr, params.SrcValidatorAddr, params.DstValidatorAddr)
	}

	redelResponses, err := RedelegationsToRedelegationResponses(ctx, k, redels)
	if err != nil {
		return nil, err
	}

	if redelResponses == nil {
		redelResponses = types.RedelegationResponses{}
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, redelResponses)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryHistoricalInfo(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryHistoricalInfoRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	hi, found := k.GetHistoricalInfo(ctx, params.Height)
	if !found {
		return nil, types.ErrNoHistoricalInfo
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, hi)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryPool(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	bondDenom := k.BondDenom(ctx)
	bondedPool := k.GetBondedPool(ctx)
	notBondedPool := k.GetNotBondedPool(ctx)

	if bondedPool == nil || notBondedPool == nil {
		return nil, errors.New("pool accounts haven't been set")
	}

	pool := types.NewPool(
		k.bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount,
		k.bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount,
	)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, pool)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryParameters(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// ______________________________________________________
// util

func DelegationToDelegationResponse(ctx sdk.Context, k Keeper, del types.Delegation) (types.DelegationResponse, error) {
	val, found := k.GetValidator(ctx, del.GetValidatorAddr())
	if !found {
		return types.DelegationResponse{}, types.ErrNoValidatorFound
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(del.DelegatorAddress)
	if err != nil {
		return types.DelegationResponse{}, err
	}

	return types.NewDelegationResp(
		delegatorAddress,
		del.GetValidatorAddr(),
		del.Shares,
		sdk.NewCoin(k.BondDenom(ctx), val.TokensFromShares(del.Shares).TruncateInt()),
	), nil
}

func DelegationsToDelegationResponses(
	ctx sdk.Context, k Keeper, delegations types.Delegations,
) (types.DelegationResponses, error) {
	resp := make(types.DelegationResponses, len(delegations))

	for i, del := range delegations {
		delResp, err := DelegationToDelegationResponse(ctx, k, del)
		if err != nil {
			return nil, err
		}

		resp[i] = delResp
	}

	return resp, nil
}

func RedelegationsToRedelegationResponses(
	ctx sdk.Context, k Keeper, redels types.Redelegations,
) (types.RedelegationResponses, error) {
	resp := make(types.RedelegationResponses, len(redels))

	for i, redel := range redels {
		valSrcAddr, err := sdk.ValAddressFromBech32(redel.ValidatorSrcAddress)
		if err != nil {
			panic(err)
		}
		valDstAddr, err := sdk.ValAddressFromBech32(redel.ValidatorDstAddress)
		if err != nil {
			panic(err)
		}

		delegatorAddress, err := sdk.AccAddressFromBech32(redel.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		val, found := k.GetValidator(ctx, valDstAddr)
		if !found {
			return nil, types.ErrNoValidatorFound
		}

		entryResponses := make([]types.RedelegationEntryResponse, len(redel.Entries))
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
			delegatorAddress,
			valSrcAddr,
			valDstAddr,
			entryResponses,
		)
	}

	return resp, nil
}
