package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	keep "github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// nolint
const (
	QueryValidators          = "validators"
	QueryValidator           = "validator"
	QueryDelegator           = "delegator"
	QueryDelegation          = "delegation"
	QueryUnbondingDelegation = "unbondingDelegation"
	QueryDelegatorValidators = "delegatorValidators"
	QueryDelegatorValidator  = "delegatorValidator"
	QueryPool                = "pool"
	QueryParameters          = "parameters"
)

// creates a querier for staking REST endpoints
func NewQuerier(k keep.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryValidators:
			return queryValidators(ctx, path[1:], k)
		case QueryValidator:
			return queryValidator(ctx, path[1:], req, k)
		case QueryDelegator:
			return queryDelegator(ctx, path[1:], req, k)
		case QueryDelegation:
			return queryDelegation(ctx, path[1:], req, k)
		case QueryUnbondingDelegation:
			return queryUnbondingDelegation(ctx, path[1:], req, k)
		case QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, path[1:], req, k)
		case QueryDelegatorValidator:
			return queryDelegatorValidator(ctx, path[1:], req, k)
		case QueryPool:
			return queryPool(ctx, k)
		case QueryParameters:
			return queryParameters(ctx, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown stake query endpoint")
		}
	}
}

// defines the params for the following queries:
// - 'custom/stake/delegator'
// - 'custom/stake/delegatorValidators'
type QueryDelegatorParams struct {
	DelegatorAddr sdk.AccAddress
}

// defines the params for the following queries:
// - 'custom/stake/validator'
type QueryValidatorParams struct {
	ValidatorAddr sdk.ValAddress
}

// defines the params for the following queries:
// - 'custom/stake/delegation'
// - 'custom/stake/unbondingDelegation'
// - 'custom/stake/delegatorValidator'
type QueryBondsParams struct {
	DelegatorAddr sdk.AccAddress
	ValidatorAddr sdk.ValAddress
}

func queryValidators(ctx sdk.Context, path []string, k keep.Keeper) (res []byte, err sdk.Error) {
	validators := k.GetBechValidators(ctx)

	res, errRes := wire.MarshalJSONIndent(k.Codec(), validators)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryValidator(ctx sdk.Context, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryValidatorParams

	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownAddress(fmt.Sprintf("incorrectly formatted request address: %s", err.Error()))
	}

	validator, found := k.GetBechValidator(ctx, params.ValidatorAddr)
	if !found {
		return []byte{}, ErrNoValidatorFound(DefaultCodespace)
	}

	res, errRes = wire.MarshalJSONIndent(k.Codec(), validator)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

// TODO query with limit
func queryDelegator(ctx sdk.Context, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryDelegatorParams
	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownAddress(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}
	delegations := k.GetDelegatorDelegationsWithoutRat(ctx, params.DelegatorAddr)
	unbondingDelegations := k.GetDelegatorUnbondingDelegations(ctx, params.DelegatorAddr)
	redelegations := k.GetRedelegations(ctx, params.DelegatorAddr)

	summary := types.DelegationSummary{
		Delegations:          delegations,
		UnbondingDelegations: unbondingDelegations,
		Redelegations:        redelegations,
	}

	res, errRes = wire.MarshalJSONIndent(k.Codec(), summary)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

// TODO query with limit
func queryDelegatorValidators(ctx sdk.Context, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryDelegatorParams

	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownAddress(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}

	validators := k.GetDelegatorBechValidators(ctx, params.DelegatorAddr)

	res, errRes = wire.MarshalJSONIndent(k.Codec(), validators)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryDelegatorValidator(ctx sdk.Context, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}

	validator := k.GetDelegatorBechValidator(ctx, params.DelegatorAddr, params.ValidatorAddr)

	res, errRes = wire.MarshalJSONIndent(k.Codec(), validator)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryDelegation(ctx sdk.Context, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}

	delegation, found := k.GetDelegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if !found {
		return []byte{}, ErrNoDelegation(DefaultCodespace)
	}

	outputDelegation := types.NewDelegationWithoutDec(delegation)
	res, errRes = wire.MarshalJSONIndent(k.Codec(), outputDelegation)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryUnbondingDelegation(ctx sdk.Context, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}

	unbond, found := k.GetUnbondingDelegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if !found {
		return []byte{}, ErrNoUnbondingDelegation(DefaultCodespace)
	}

	res, errRes = wire.MarshalJSONIndent(k.Codec(), unbond)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryPool(ctx sdk.Context, k keep.Keeper) (res []byte, err sdk.Error) {
	pool := k.GetPool(ctx)

	res, errRes := wire.MarshalJSONIndent(k.Codec(), pool)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryParameters(ctx sdk.Context, k keep.Keeper) (res []byte, err sdk.Error) {
	params := k.GetParams(ctx)

	res, errRes := wire.MarshalJSONIndent(k.Codec(), params)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}
