package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ types.QueryServer = Keeper{}

// Connection implements the Query/Connection gRPC method
func (q Keeper) Connection(c context.Context, req *types.QueryConnectionRequest) (*types.QueryConnectionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := host.ConnectionIdentifierValidator(req.ConnectionID); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	connection, found := q.GetConnection(ctx, req.ConnectionID)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrap(types.ErrConnectionNotFound, req.ConnectionID).Error(),
		)
	}

	return &types.QueryConnectionResponse{
		Connection:  &connection,
		ProofHeight: uint64(ctx.BlockHeight()),
	}, nil
}

// Connections implements the Query/Connections gRPC method
func (q Keeper) Connections(c context.Context, req *types.QueryConnectionsRequest) (*types.QueryConnectionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	connections := []*types.IdentifiedConnection{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), host.KeyConnectionPrefix)

	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		var result types.ConnectionEnd
		if err := q.cdc.UnmarshalBinaryBare(value, &result); err != nil {
			return err
		}

		connectionID, err := host.ParseConnectionPath(string(key))
		if err != nil {
			return err
		}

		identifiedConnection := types.NewIdentifiedConnection(connectionID, result)
		connections = append(connections, &identifiedConnection)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryConnectionsResponse{
		Connections: connections,
		Pagination:  pageRes,
		Height:      ctx.BlockHeight(),
	}, nil
}

// ClientConnections implements the Query/ClientConnections gRPC method
func (q Keeper) ClientConnections(c context.Context, req *types.QueryClientConnectionsRequest) (*types.QueryClientConnectionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := host.ClientIdentifierValidator(req.ClientID); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	clientConnectionPaths, found := q.GetClientConnectionPaths(ctx, req.ClientID)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrap(types.ErrClientConnectionPathsNotFound, req.ClientID).Error(),
		)
	}

	return &types.QueryClientConnectionsResponse{
		ConnectionPaths: clientConnectionPaths,
		ProofHeight:     uint64(ctx.BlockHeight()),
	}, nil
}

// ConnectionClientState implements the Query/ConnectionClientState gRPC method
func (q Keeper) ConnectionClientState(c context.Context, req *types.QueryConnectionClientStateRequest) (*types.QueryConnectionClientStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := host.ConnectionIdentifierValidator(req.ConnectionID); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	connection, found := q.GetConnection(ctx, req.ConnectionID)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(types.ErrConnectionNotFound, "connection-id: %s", req.ConnectionID).Error(),
		)
	}

	clientState, found := q.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(clienttypes.ErrClientNotFound, "client-id: %s", connection.ClientID).Error(),
		)
	}

	identifiedClientState := clienttypes.NewIdentifiedClientState(connection.ClientID, clientState)

	return types.NewQueryConnectionClientStateResponse(identifiedClientState, nil, ctx.BlockHeight()), nil

}

// ConnectionConsensusState implements the Query/ConnectionConsensusState gRPC method
func (q Keeper) ConnectionConsensusState(c context.Context, req *types.QueryConnectionConsensusStateRequest) (*types.QueryConnectionConsensusStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := host.ConnectionIdentifierValidator(req.ConnectionID); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	connection, found := q.GetConnection(ctx, req.ConnectionID)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(types.ErrConnectionNotFound, "connection-id: %s", req.ConnectionID).Error(),
		)
	}

	consensusState, found := q.clientKeeper.GetClientConsensusState(ctx, connection.ClientID, req.Height)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(clienttypes.ErrConsensusStateNotFound, "client-id: %s", connection.ClientID).Error(),
		)
	}

	anyConsensusState, err := clienttypes.PackConsensusState(consensusState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return types.NewQueryConnectionConsensusStateResponse(connection.ClientID, anyConsensusState, consensusState.GetHeight(), nil, ctx.BlockHeight()), nil
}
