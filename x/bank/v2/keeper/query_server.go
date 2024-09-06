package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/x/bank/v2/types"
)

var _ types.QueryServer = querier{}

type querier struct {
	*Keeper
}

// NewMsgServer creates a new bank/v2 query server.
func NewQuerier(k *Keeper) types.QueryServer {
	return querier{k}
}

// Params implements types.QueryServer.
func (q querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	params, err := q.params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}
