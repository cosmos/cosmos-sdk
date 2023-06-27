package types

import (
	context "context"
)

type ConsensusKeeper interface {
	Params(ctx context.Context, _ *QueryParamsRequest) (*QueryParamsResponse, error)
}
