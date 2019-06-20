package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/uniswap/internal/types"
)

// NewQuerier creates a querier for uniswap REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryBalance:
			return queryBalance(ctx, req, k)
		case types.QueryLiquidity:
			return queryLiquidity(ctx, req, k)
		case types.QueryParameters:
			return queryParameters(ctx, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
		}
	}
}

// queryBalance returns the provided addresses UNI balance upon success
// or an error if the query fails.
func queryBalance(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var address sdk.AccAddress
	err := k.cdc.UnmarshalJSON(req.Data, &address)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	balance := k.GetUNIForAddress(ctx, address)

	bz, err := k.cdc.MarshalJSONIndent(balance, "", " ")
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// queryLiquidity returns the total liquidity avaliable for the provided denomination
// upon success or an error if the query fails.
func queryLiquidity(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var denom string
	err := k.cdc.UnmarshalJSON(req.Data, &denom)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	liquidity := k.GetReservePool(ctx, denom)

	bz, err := k.cdc.MarshalJSONIndent(liquidity, "", " ")
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// queryParameters returns uniswap module parameter queried for upon success
// or an error if the query fails
func queryParameters(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	switch path[0] {
	case types.ParamFee:
		bz, err := k.cdc.MarshalJSONIndent(k.GetFeeParam(ctx), "", " ")
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	case types.ParamNativeDenom:
		bz, err := k.cdc.MarshalJSONIndent(k.GetNativeDenom(ctx), "", " ")
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	default:
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}
}
