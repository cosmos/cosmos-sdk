package slashing

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the slashing querier
const (
	QueryParameters   = "parameters"
	QuerySigningInfo  = "signingInfo"
	QuerySigningInfos = "signingInfos"
)

// NewQuerier creates a new querier for slashing clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryParameters:
			return queryParams(ctx, k)
		case QuerySigningInfo:
			return querySigningInfo(ctx, req, k)
		case QuerySigningInfos:
			return querySigningInfos(ctx, req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown staking query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(moduleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}

// QuerySigningInfoParams defines the params for the following queries:
// - 'custom/slashing/signingInfo'
type QuerySigningInfoParams struct {
	ConsAddress sdk.ConsAddress
}

func NewQuerySigningInfoParams(consAddr sdk.ConsAddress) QuerySigningInfoParams {
	return QuerySigningInfoParams{consAddr}
}

func querySigningInfo(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QuerySigningInfoParams

	err := moduleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	signingInfo, found := k.getValidatorSigningInfo(ctx, params.ConsAddress)
	if !found {
		return nil, ErrNoSigningInfoFound(DefaultCodespace, params.ConsAddress)
	}

	res, err := codec.MarshalJSONIndent(moduleCdc, signingInfo)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

// QuerySigningInfosParams defines the params for the following queries:
// - 'custom/slashing/signingInfos'
type QuerySigningInfosParams struct {
	Page, Limit int
}

func NewQuerySigningInfosParams(page, limit int) QuerySigningInfosParams {
	return QuerySigningInfosParams{page, limit}
}

func querySigningInfos(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QuerySigningInfosParams

	err := moduleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	if params.Limit == 0 {
		// set the default limit to max bonded if no limit was provided
		params.Limit = int(k.validatorSet.MaxValidators(ctx))
	}

	var signingInfos []ValidatorSigningInfo

	k.IterateValidatorSigningInfos(ctx, func(consAddr sdk.ConsAddress, info ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	})

	// get pagination bounds
	start := (params.Page - 1) * params.Limit
	end := params.Limit + start
	if end >= len(signingInfos) {
		end = len(signingInfos)
	}

	if start >= len(signingInfos) {
		// page is out of bounds
		signingInfos = []ValidatorSigningInfo{}
	} else {
		signingInfos = signingInfos[start:end]
	}

	res, err := codec.MarshalJSONIndent(moduleCdc, signingInfos)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}
