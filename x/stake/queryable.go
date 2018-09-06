package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	keep "github.com/cosmos/cosmos-sdk/x/stake/keeper"
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
func NewQuerier(k keep.Keeper, cdc *wire.Codec) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryValidators:
			return queryValidators(ctx, cdc, path[1:], k)
		case QueryValidator:
			return queryValidator(ctx, cdc, path[1:], req, k)
		case QueryDelegator:
			return queryDelegator(ctx, cdc, path[1:], req, k)
		case QueryDelegation:
			return queryDelegation(ctx, cdc, path[1:], req, k)
		case QueryUnbondingDelegation:
			return queryUnbondingDelegation(ctx, cdc, path[1:], req, k)
		case QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, cdc, path[1:], req, k)
		case QueryDelegatorValidator:
			return queryDelegatorValidator(ctx, cdc, path[1:], req, k)
		case QueryPool:
			return queryPool(ctx, cdc, k)
		case QueryParameters:
			return queryParameters(ctx, cdc, k)
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

func queryValidators(ctx sdk.Context, cdc *wire.Codec, path []string, k keep.Keeper) (res []byte, err sdk.Error) {
	validators := k.GetValidators(ctx)

	res, errRes := wire.MarshalJSONIndent(cdc, validators)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryValidator(ctx sdk.Context, cdc *wire.Codec, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryValidatorParams

	errRes := cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownAddress(fmt.Sprintf("incorrectly formatted request address: %s", err.Error()))
	}

	validator, found := k.GetValidator(ctx, params.ValidatorAddr)
	if !found {
		return []byte{}, ErrNoValidatorFound(DefaultCodespace)
	}

	bechValidator, errRes := validator.Bech32Validator()
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not bech32ify validator: %s", errRes.Error()))
	}

	res, errRes = wire.MarshalJSONIndent(cdc, bechValidator)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

// TODO query with limit
func queryDelegator(ctx sdk.Context, cdc *wire.Codec, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryDelegatorParams
	errRes := cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownAddress(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}
	// delegations := k.GetDelegatorDelegations(ctx, params.DelegatorAddr)
	// unbondingDelegations := k.GetUnbondingDelegations(ctx, params.DelegatorAddr)
	// redelegations := k.GetRedelegations(ctx, params.DelegatorAddr)
	//
	// summary := types.DelegationSummary{
	// 	Delegations:          delegations,
	// 	UnbondingDelegations: unbondingDelegations,
	// 	Redelegations:        redelegations,
	// }
	//
	// res, errRes = wire.MarshalJSONIndent(cdc, summary)
	// if errRes != nil {
	// 	return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	// }
	return res, nil
}

// TODO query with limit
func queryDelegatorValidators(ctx sdk.Context, cdc *wire.Codec, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryDelegatorParams

	errRes := cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownAddress(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}

	validators := k.GetDelegatorBechValidators(ctx, params.DelegatorAddr)

	res, errRes = wire.MarshalJSONIndent(cdc, validators)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryDelegatorValidator(ctx sdk.Context, cdc *wire.Codec, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}

	validator := k.GetDelegatorValidator(ctx, params.DelegatorAddr, params.ValidatorAddr)

	bechValidator, errRes := validator.Bech32Validator()
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not bech32ify validator: %s", errRes.Error()))
	}

	res, errRes = wire.MarshalJSONIndent(cdc, bechValidator)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryDelegation(ctx sdk.Context, cdc *wire.Codec, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}

	delegation, found := k.GetDelegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if !found {
		return []byte{}, ErrNoDelegation(DefaultCodespace)
	}

	delegationREST := delegation.ToRest()
	res, errRes = wire.MarshalJSONIndent(cdc, delegationREST)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryUnbondingDelegation(ctx sdk.Context, cdc *wire.Codec, path []string, req abci.RequestQuery, k keep.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams

	errRes := cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request address: %s", errRes.Error()))
	}

	unbond, found := k.GetUnbondingDelegation(ctx, params.DelegatorAddr, params.ValidatorAddr)
	if !found {
		return []byte{}, ErrNoUnbondingDelegation(DefaultCodespace)
	}

	res, errRes = wire.MarshalJSONIndent(cdc, unbond)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryPool(ctx sdk.Context, cdc *wire.Codec, k keep.Keeper) (res []byte, err sdk.Error) {
	pool := k.GetPool(ctx)

	res, errRes := wire.MarshalJSONIndent(cdc, pool)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}

func queryParameters(ctx sdk.Context, cdc *wire.Codec, k keep.Keeper) (res []byte, err sdk.Error) {
	params := k.GetParams(ctx)

	res, errRes := wire.MarshalJSONIndent(cdc, params)
	if errRes != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", errRes.Error()))
	}
	return res, nil
}
