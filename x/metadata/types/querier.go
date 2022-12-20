package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// query endpoints supported by the auth Querier
const (
	QueryParams = "params"
)

var _ QueryServer = &GrpcQuerier{}

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
func (g GrpcQuerier) Params(stdCtx context.Context, _ *QueryParamsRequest) (*QueryParamsResponse, error) {
	var params Params
	ctx := sdk.UnwrapSDKContext(stdCtx)
	if g.paramSource.Has(ctx, ParamStoreKeyMetadataPrices) {
		g.paramSource.Get(ctx, ParamStoreKeyMetadataPrices, &params)
	} else {
		params = DefaultParams()
	}
	return &QueryParamsResponse{
		Params: params,
	}, nil
}
