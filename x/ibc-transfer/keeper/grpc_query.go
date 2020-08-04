package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

var _ types.QueryServer = Keeper{}

// DenomTrace implements the Query/DenomTrace gRPC method
func (q Keeper) DenomTrace(c context.Context, req *types.QueryDenomTraceRequest) (*types.QueryDenomTraceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	hash, err := types.ParseHexHash(req.Hash)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid denom trace hash %s, %s", req.Hash, err))
	}

	ctx := sdk.UnwrapSDKContext(c)
	denomTrace, found := q.GetDenomTrace(ctx, hash)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrap(types.ErrInvalidDenomForTransfer, req.Hash).Error(), // TODO: update error
		)
	}

	return &types.QueryDenomTraceResponse{
		DenomTrace: &denomTrace,
	}, nil
}

// DenomTraces implements the Query/DenomTraces gRPC method
func (q Keeper) DenomTraces(c context.Context, req *types.QueryDenomTracesRequest) (*types.QueryDenomTracesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	traces := types.Traces{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), types.DenomTraceKey)

	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		var result types.DenomTrace
		if err := q.cdc.UnmarshalBinaryBare(value, &result); err != nil {
			return err
		}

		traces = append(traces, result)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryDenomTracesResponse{
		DenomTraces: traces,
		Pagination:  pageRes,
	}, nil
}
