package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewQuerier creates a querier for uniswap REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryBalance:
			return queryBalance(ctx, req, k)
		case types.QueryLiquidity:
			return queryLiquidity(ctx, req, k)
		case types.Parameters:
			return queryLiquidity(ctx, req, k)
		default:
			return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
		}
	}
}

// queryBalance returns the provided addresses UNI balance upon success
// or an error if the query fails.
func queryBalance(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryBalance
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	balance := k.GetUNIForAddress(params.Address)

	bz, err := codec.MarshalJSONIndent(k.cdc, balance)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// queryLiquidity returns the total liquidity avaliable at the provided exchange
// upon success or an error if the query fails.
func queryLiquidity(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryLiquidity
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	liquidity := k.GetExchange(ctx, params.Denom)

	bz, err := codec.MarshalJSONIndent(k.cdc, liquidity)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// queryParameters returns uniswap module parameter queried for upon success
// or an error if the query fails
func queryParameters(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	switch path[0] {
	case ParamFee:
		bz, err := codec.MarshalJSONIndent(k.cdc, k.GetFeeParams(ctx))
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	case ParamNativeAsset:
		bz, err := codec.MarshalJSONIndent(k.cdc, k.GetNativeAsset())
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	default:
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}
}
