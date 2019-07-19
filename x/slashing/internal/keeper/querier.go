package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

// NewQuerier creates a new querier for slashing clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryParameters:
			return queryParams(ctx, k)
		case types.QuerySigningInfo:
			return querySigningInfo(ctx, req, k)
		case types.QuerySigningInfos:
			return querySigningInfos(ctx, req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown staking query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}

func querySigningInfo(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QuerySigningInfoParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	signingInfo, found := k.GetValidatorSigningInfo(ctx, params.ConsAddress)
	if !found {
		return nil, types.ErrNoSigningInfoFound(types.DefaultCodespace, params.ConsAddress)
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, signingInfo)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

func querySigningInfos(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QuerySigningInfosParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	var signingInfos []types.ValidatorSigningInfo

	k.IterateValidatorSigningInfos(ctx, func(consAddr sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	})

	start, end := client.Paginate(len(signingInfos), params.Page, params.Limit, int(k.sk.MaxValidators(ctx)))
	if start < 0 || end < 0 {
		signingInfos = []types.ValidatorSigningInfo{}
	} else {
		signingInfos = signingInfos[start:end]
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, signingInfos)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}
