package querier

import (
	"fmt"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// query endpoints supported by the staking Querier
const (
	QueryValidators                    = "validators"
	QueryValidator                     = "validator"
	QueryDelegatorDelegations          = "delegatorDelegations"
	QueryDelegatorUnbondingDelegations = "delegatorUnbondingDelegations"
	QueryRedelegations                 = "redelegations"
	QueryValidatorDelegations          = "validatorDelegations"
	QueryValidatorRedelegations        = "validatorRedelegations"
	QueryValidatorUnbondingDelegations = "validatorUnbondingDelegations"
	QueryDelegation                    = "delegation"
	QueryUnbondingDelegation           = "unbondingDelegation"
	QueryDelegatorValidators           = "delegatorValidators"
	QueryDelegatorValidator            = "delegatorValidator"
	QueryPool                          = "pool"
	QueryParameters                    = "parameters"
)

// creates a querier for staking REST endpoints
func NewQuerier(k keep.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryValidators:
			return queryValidators(ctx, req, k)
		case QueryValidator:
			return queryValidator(ctx, req, k)
		case QueryValidatorDelegations:
			return queryValidatorDelegations(ctx, req, k)
		case QueryValidatorUnbondingDelegations:
			return queryValidatorUnbondingDelegations(ctx, req, k)
		case QueryDelegation:
			return queryDelegation(ctx, req, k)
		case QueryUnbondingDelegation:
			return queryUnbondingDelegation(ctx, req, k)
		case QueryDelegatorDelegations:
			return queryDelegatorDelegations(ctx, req, k)
		case QueryDelegatorUnbondingDelegations:
			return queryDelegatorUnbondingDelegations(ctx, req, k)
		case QueryRedelegations:
			return queryRedelegations(ctx, req, k)
		case QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, req, k)
		case QueryDelegatorValidator:
			return queryDelegatorValidator(ctx, req, k)
		case QueryPool:
			return queryPool(ctx, k)
		case QueryParameters:
			return queryParameters(ctx, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown staking query endpoint")
		}
	}
}

// defines the params for the following queries:
// - 'custom/staking/delegatorDelegations'
// - 'custom/staking/delegatorUnbondingDelegations'
// - 'custom/staking/delegatorRedelegations'
// - 'custom/staking/delegatorValidators'
type QueryDelegatorParams struct {
	DelegatorAddr sdk.AccAddress
}

func NewQueryDelegatorParams(delegatorAddr sdk.AccAddress) QueryDelegatorParams {
	return QueryDelegatorParams{
		DelegatorAddr: delegatorAddr,
	}
}

// defines the params for the following queries:
// - 'custom/staking/validator'
// - 'custom/staking/validatorDelegations'
// - 'custom/staking/validatorUnbondingDelegations'
// - 'custom/staking/validatorRedelegations'
type QueryValidatorParams struct {
	ValidatorAddr sdk.ValAddress
}

func NewQueryValidatorParams(validatorAddr sdk.ValAddress) QueryValidatorParams {
	return QueryValidatorParams{
		ValidatorAddr: validatorAddr,
	}
}

// defines the params for the following queries:
// - 'custom/staking/delegation'
// - 'custom/staking/unbondingDelegation'
// - 'custom/staking/delegatorValidator'
type QueryBondsParams struct {
	DelegatorAddr sdk.AccAddress
	ValidatorAddr sdk.ValAddress
}

func NewQueryBondsParams(delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) QueryBondsParams {
	return QueryBondsParams{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
	}
}

// defines the params for the following queries:
// - 'custom/staking/redelegation'
type QueryRedelegationParams struct {
	DelegatorAddr    sdk.AccAddress
	SrcValidatorAddr sdk.ValAddress
	DstValidatorAddr sdk.ValAddress
}

func NewQueryRedelegationParams(delegatorAddr sdk.AccAddress, srcValidatorAddr sdk.ValAddress, dstValidatorAddr sdk.ValAddress) QueryRedelegationParams {
	return QueryRedelegationParams{
		DelegatorAddr:    delegatorAddr,
		SrcValidatorAddr: srcValidatorAddr,
		DstValidatorAddr: dstValidatorAddr,
	}
}

func queryValidators(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) ([]byte, sdk.Error) {
	var params QueryValidatorsParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	stakingParams := k.GetParams(ctx)
	if params.Limit == 0 {
		params.Limit = int(stakingParams.MaxValidators)
	}

	validators := k.GetAllValidators(ctx)
	filteredVals := make([]types.Validator, 0, len(validators))

	for _, val := range validators {
		if strings.ToLower(val.GetStatus().String()) == strings.ToLower(params.Status) {
			filteredVals = append(filteredVals, val)
		}
	}

	// get pagination bounds
	start := (params.Page - 1) * params.Limit
	end := params.Limit + start
	if end >= len(filteredVals) {
		end = len(filteredVals)
	}

	if start >= len(filteredVals) {
		// page is out of bounds
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

func queryValidator(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryValidatorParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	validator, found := k.GetValidator(ctx, params.ValidatorAddr)
	if !found {
		return []byte{}, types.ErrNoValidatorFound(types.DefaultCodespace)
	}

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, validator)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errRes.Error()))
	}
	return res, nil
}

func queryValidatorDelegations(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryValidatorParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	delegations := k.GetValidatorDelegations(ctx, params.ValidatorAddr)
	delegationResps, err := delegationsToDelegationResponses(ctx, k, delegations)
	if err != nil {
		return nil, err
	}

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, delegationResps)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", errRes.Error()))
	}

	return res, nil
}

func queryValidatorUnbondingDelegations(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryValidatorParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	unbonds := k.GetUnbondingDelegationsFromValidator(ctx, params.ValidatorAddr)

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, unbonds)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errRes.Error()))
	}
	return res, nil
}

func queryDelegatorDelegations(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryDelegatorParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	delegations := k.GetAllDelegatorDelegations(ctx, params.DelegatorAddr)
	delegationResps, err := delegationsToDelegationResponses(ctx, k, delegations)
	if err != nil {
		return nil, err
	}

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, delegationResps)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", errRes.Error()))
	}

	return res, nil
}

func queryDelegatorUnbondingDelegations(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryDelegatorParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	unbondingDelegations := k.GetAllUnbondingDelegations(ctx, params.DelegatorAddr)

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, unbondingDelegations)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errRes.Error()))
	}
	return res, nil
}

func queryDelegatorValidators(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryDelegatorParams

	stakingParams := k.GetParams(ctx)

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	validators := k.GetDelegatorValidators(ctx, params.DelegatorAddr, stakingParams.MaxValidators)

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, validators)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errRes.Error()))
	}
	return res, nil
}

func queryDelegatorValidator(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	validator, err := k.GetDelegatorValidator(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if err != nil {
		return
	}

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, validator)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errRes.Error()))
	}
	return res, nil
}

func queryDelegation(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	delegation, found := k.GetDelegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if !found {
		return []byte{}, types.ErrNoDelegation(types.DefaultCodespace)
	}

	delegationResp, err := delegationToDelegationResponse(ctx, k, delegation)
	if err != nil {
		return nil, err
	}

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, delegationResp)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", errRes.Error()))
	}

	return res, nil
}

func queryUnbondingDelegation(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	unbond, found := k.GetUnbondingDelegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if !found {
		return []byte{}, types.ErrNoUnbondingDelegation(types.DefaultCodespace)
	}

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, unbond)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errRes.Error()))
	}
	return res, nil
}

func queryRedelegations(ctx sdk.Context, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryRedelegationParams

	errRes := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(string(req.Data))
	}

	var redels []types.Redelegation

	if !params.DelegatorAddr.Empty() && !params.SrcValidatorAddr.Empty() && !params.DstValidatorAddr.Empty() {
		redel, found := k.GetRedelegation(ctx, params.DelegatorAddr, params.SrcValidatorAddr, params.DstValidatorAddr)
		if !found {
			return nil, types.ErrNoRedelegation(types.DefaultCodespace)
		}

		redels = []types.Redelegation{redel}
	} else if params.DelegatorAddr.Empty() && !params.SrcValidatorAddr.Empty() && params.DstValidatorAddr.Empty() {
		redels = k.GetRedelegationsFromValidator(ctx, params.SrcValidatorAddr)
	} else {
		redels = k.GetAllRedelegations(ctx, params.DelegatorAddr, params.SrcValidatorAddr, params.DstValidatorAddr)
	}

	redelResponses, err := redelegationsToRedelegationResponses(ctx, k, redels)
	if err != nil {
		return nil, err
	}

	res, errRes = codec.MarshalJSONIndent(types.ModuleCdc, redelResponses)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", errRes.Error()))
	}

	return res, nil
}

func queryPool(ctx sdk.Context, k keep.Keeper) (res []byte, err sdk.Error) {
	pool := k.GetPool(ctx)

	res, errRes := codec.MarshalJSONIndent(types.ModuleCdc, pool)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errRes.Error()))
	}
	return res, nil
}

func queryParameters(ctx sdk.Context, k keep.Keeper) (res []byte, err sdk.Error) {
	params := k.GetParams(ctx)

	res, errRes := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if errRes != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errRes.Error()))
	}
	return res, nil
}

// QueryValidatorsParams defines the params for the following queries:
// - 'custom/staking/validators'
type QueryValidatorsParams struct {
	Page, Limit int
	Status      string
}

func NewQueryValidatorsParams(page, limit int, status string) QueryValidatorsParams {
	return QueryValidatorsParams{page, limit, status}
}
