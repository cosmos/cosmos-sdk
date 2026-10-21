package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

type queryServer struct {
	k Keeper
}

var _ types.QueryServer = queryServer{}

// NewQueryServer returns an implementation of the x/safety QueryServer.
func NewQueryServer(k Keeper) types.QueryServer {
	return queryServer{k: k}
}

// Params returns the module params.
func (q queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

// DisabledList returns a page of disabled patterns.
func (q queryServer) DisabledList(ctx context.Context, req *types.QueryDisabledListRequest) (*types.QueryDisabledListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	patterns, pageRes, err := query.CollectionPaginate(ctx, q.k.DisabledList, req.Pagination,
		func(p string, _ collections.NoValue) (string, error) { return p, nil },
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &types.QueryDisabledListResponse{DisabledPatterns: patterns, Pagination: pageRes}, nil
}

// IsAllowed reports whether a msg type URL may execute.
func (q queryServer) IsAllowed(ctx context.Context, req *types.QueryIsAllowedRequest) (*types.QueryIsAllowedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := types.ValidateTypeURL(req.MsgTypeUrl); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	allowed, err := q.k.IsAllowed(ctx, req.MsgTypeUrl)
	switch {
	case errors.Is(err, types.ErrMsgDisabled):
		return &types.QueryIsAllowedResponse{Allowed: false}, nil
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryIsAllowedResponse{Allowed: allowed}, nil
}
