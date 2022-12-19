package metadata

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/metadata/types"
)

var _ types.QueryServer = &GrpcQuerier{}

// ParamSource is a read only subset of paramtypes.Subspace
type ParamSource interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Has(ctx sdk.Context, key []byte) bool
}

type GrpcQuerier struct {
	paramSource ParamSource
}

func NewGrpcQuerier(paramSource ParamSource) GrpcQuerier {
	return GrpcQuerier{paramSource: paramSource}
}

// MetadataParams return metadata params
func (g GrpcQuerier) MetadataParams(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	var params types.Params
	ctx := sdk.UnwrapSDKContext(stdCtx)
	if g.paramSource.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		g.paramSource.Get(ctx, types.ParamStoreKeyMinGasPrices, &params)
	}
	return &types.QueryParamsResponse{
		MinimumGasPrices: params,
	}, nil
}
