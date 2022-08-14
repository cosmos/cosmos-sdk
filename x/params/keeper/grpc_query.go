package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

var _ proposal.QueryServer = Keeper{}

// Params returns subspace params
func (k Keeper) Params(c context.Context, req *proposal.QueryParamsRequest) (*proposal.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Subspace == "" || req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ss, ok := k.GetSubspace(req.Subspace)
	if !ok {
		return nil, sdkerrors.Wrap(proposal.ErrUnknownSubspace, req.Subspace)
	}

	ctx := sdk.UnwrapSDKContext(c)
	rawValue := ss.GetRaw(ctx, []byte(req.Key))
	param := proposal.NewParamChange(req.Subspace, req.Key, string(rawValue))

	return &proposal.QueryParamsResponse{Param: param}, nil
}
