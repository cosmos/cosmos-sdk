package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
)

// TODO Redelegations
func NewQuerier(k keeper.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case "validators":
			return queryValidators(ctx, path[1:], req, k)
		case "validator":
			return queryValidator(ctx, path[1:], req, k)
		// case "delegator":
		// 	return queryDelegator(ctx, path[1:], req, k)
		case "delegatorValidators":
			return queryDelegatorValidators(ctx, path[1:], req, k)
		case "delegatorValidator":
			return queryDelegatorValidator(ctx, path[1:], req, k)
		case "delegation":
			return queryDelegation(ctx, path[1:], req, k)
		case "unbonding-delegation":
			return queryUnbondingDelegation(ctx, path[1:], req, k)
		case "pool":
			return queryPool(ctx, path[1:], req, k)
		case "parameters":
			return queryParameters(ctx, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown stake query endpoint")
		}
	}
}

// Params for queries:
// - 'custom/stake/delegator'
// - 'custom/stake/delegator/txs'
// - 'custom/stake/delegator/validators'
// - 'custom/stake/validator'
type QueryAddressParams struct {
	accountAddr sdk.AccAddress
}

// Params for queries
// - 'custom/stake/delegator/delegations'
// - 'custom/stake/delegator/unbonding_delegations'
// - 'custom/stake/delegator/validator'
type QueryBondsParams struct {
	delegatorAddr sdk.AccAddress
	validatorAddr sdk.AccAddress
}

func queryValidators(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	validators := k.GetAllValidators(ctx)

	res, errRes := wire.MarshalJSONIndent(k.Codec(), validators)
	if err != nil {
		panic(errRes.Error())
	}
	return res, nil
}

func queryValidator(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	var params QueryAddressParams
	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data: \n%s", err.Error()))
	}

	validator, found := k.GetValidator(ctx, params.accountAddr)
	if !found {
		return []byte{}, ErrNoValidatorFound(DefaultCodespace)
	}

	res, errRes = wire.MarshalJSONIndent(k.Codec(), validator)
	if errRes != nil {
		panic(fmt.Sprintf("could not marshal result to JSON:\n%s", errRes.Error()))
	}
	return res, nil
}

// TODO query with limit
// func queryDelegator(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
// 	var params QueryAddressParams
// 	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
// 	if errRes != nil {
// 		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", errRes.Error()))
// 	}
//
// 	res, errRes = wire.MarshalJSONIndent(k.Codec(), deposit)
// 	if errRes != nil {
// 		panic("could not marshal result to JSON")
// 	}
// 	return res, nil
// }

// TODO query with limit
func queryDelegatorValidators(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	var params QueryAddressParams
	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data :\n%s", errRes.Error()))
	}

	validators := k.GetDelegatorValidators(ctx, params.accountAddr)

	res, errRes = wire.MarshalJSONIndent(k.Codec(), validators)
	if errRes != nil {
		panic(fmt.Sprintf("could not marshal result to JSON:\n%s", errRes.Error()))
	}
	return res, nil
}

func queryDelegatorValidator(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams
	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data :\n%s", errRes.Error()))
	}

	validator := k.GetDelegatorValidator(ctx, params.delegatorAddr, params.validatorAddr)

	res, errRes = wire.MarshalJSONIndent(k.Codec(), validator)
	if errRes != nil {
		panic(fmt.Sprintf("could not marshal result to JSON:\n%s", errRes.Error()))
	}
	return res, nil
}

func queryDelegation(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams
	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data :\n%s", errRes.Error()))
	}

	delegation, found := k.GetDelegation(ctx, params.delegatorAddr, params.validatorAddr)
	if !found {
		return []byte{}, ErrNoDelegation(DefaultCodespace)
	}

	res, errRes = wire.MarshalJSONIndent(k.Codec(), delegation)
	if errRes != nil {
		panic(fmt.Sprintf("could not marshal result to JSON:\n%s", errRes.Error()))
	}
	return res, nil
}

func queryUnbondingDelegation(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	var params QueryBondsParams
	errRes := k.Codec().UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data :\n%s", errRes.Error()))
	}

	unbond, found := k.GetUnbondingDelegation(ctx, params.delegatorAddr, params.validatorAddr)
	if !found {
		return []byte{}, ErrNoUnbondingDelegation(DefaultCodespace)
	}

	res, errRes = wire.MarshalJSONIndent(k.Codec(), unbond)
	if errRes != nil {
		panic(fmt.Sprintf("could not marshal result to JSON:\n%s", errRes.Error()))
	}
	return res, nil
}

func queryPool(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	pool := k.GetPool(ctx)

	res, errRes := wire.MarshalJSONIndent(k.Codec(), pool)
	if errRes != nil {
		panic(fmt.Sprintf("could not marshal result to JSON:\n%s", errRes.Error()))
	}
	return res, nil
}

func queryParameters(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	params := k.GetParams(ctx)

	res, errRes := wire.MarshalJSONIndent(k.Codec(), params)
	if errRes != nil {
		panic(fmt.Sprintf("could not marshal result to JSON:\n%s", errRes.Error()))
	}
	return res, nil
}
