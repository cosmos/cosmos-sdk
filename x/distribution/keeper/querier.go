package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryParams              = "params"
	QueryOutstandingRewards  = "outstanding_rewards"
	QueryValidatorCommission = "validator_commission"
	QueryValidatorSlashes    = "validator_slashes"
	QueryDelegationInfo      = "delegation_info"
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
		case QueryDelegationInfo:
			return queryDelegationInfo(ctx, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown distr query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	bz, err := codec.MarshalJSONIndent(k.cdc, k.GetParams(ctx))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryOutstandingRewards(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	bz, err := codec.MarshalJSONIndent(k.cdc, k.GetOutstandingRewards(ctx))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryValidatorCommission(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	return []byte{}, nil
}

func queryValidatorSlashes(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	return []byte{}, nil
}

func queryDelegationInfo(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	return []byte{}, nil
}
