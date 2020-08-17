package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ types.QueryServer = Keeper{}

// ClientState implements the Query/ClientState gRPC method
func (q Keeper) ClientState(c context.Context, req *types.QueryClientStateRequest) (*types.QueryClientStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := host.ClientIdentifierValidator(req.ClientId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	clientState, found := q.GetClientState(ctx, req.ClientId)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrap(types.ErrClientNotFound, req.ClientId).Error(),
		)
	}

	any, err := types.PackClientState(clientState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryClientStateResponse{
		ClientState: any,
		ProofHeight: uint64(ctx.BlockHeight()),
	}, nil
}

// ClientStates implements the Query/ClientStates gRPC method
func (q Keeper) ClientStates(c context.Context, req *types.QueryClientStatesRequest) (*types.QueryClientStatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	clientStates := []*types.IdentifiedClientState{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), host.KeyConnectionPrefix)

	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {

		clientState, err := q.UnmarshalClientState(value)
		if err != nil {
			return err
		}

		clientID, err := host.ParseClientPath(string(key))
		if err != nil {
			return err
		}

		identifiedConnection := types.NewIdentifiedClientState(clientID, clientState)
		clientStates = append(clientStates, &identifiedConnection)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryClientStatesResponse{
		ClientStates: clientStates,
		Pagination:   pageRes,
	}, nil
}

// ConsensusState implements the Query/ConsensusState gRPC method
func (q Keeper) ConsensusState(c context.Context, req *types.QueryConsensusStateRequest) (*types.QueryConsensusStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := host.ClientIdentifierValidator(req.ClientId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Height == 0 {
		return nil, status.Error(codes.InvalidArgument, "consensus state height cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	consensusState, found := q.GetClientConsensusState(ctx, req.ClientId, req.Height)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(types.ErrConsensusStateNotFound, "client-id: %s, height: %d", req.ClientId, req.Height).Error(),
		)
	}

	any, err := types.PackConsensusState(consensusState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryConsensusStateResponse{
		ConsensusState: any,
		ProofHeight:    uint64(ctx.BlockHeight()),
	}, nil
}
