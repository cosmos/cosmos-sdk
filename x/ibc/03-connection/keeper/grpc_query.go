package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ types.QueryServer = Keeper{}

// Connection implements the Query/Connection gRPC method
func (q Keeper) Connection(c context.Context, req *types.QueryConnectionRequest) (*types.QueryConnectionResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
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

	return &types.QueryConnectionResponse{Connection: &connection}, nil
}

// ClientConnections implements the Query/ClientConnections gRPC method
func (q Keeper) ClientConnections(c context.Context, req *types.QueryClientConnectionsRequest) (*types.QueryClientConnectionsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
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
	}, nil
}

// // AllBalances implements the Query/AllBalances gRPC method
// func (q Keeper) AllBalances(c context.Context, req *types.QueryAllBalancesRequest) (*types.QueryAllBalancesResponse, error) {
// 	if req == nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "empty request")
// 	}

// 	if len(req.Address) == 0 {
// 		return nil, status.Errorf(codes.InvalidArgument, "invalid address")
// 	}

// 	ctx := sdk.UnwrapSDKContext(c)
// 	balances := q.GetAllBalances(ctx, req.Address)

// 	return &types.QueryAllBalancesResponse{Balances: balances}, nil
// }

// // TotalSupply implements the Query/TotalSupply gRPC method
// func (q Keeper) TotalSupply(c context.Context, _ *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(c)
// 	totalSupply := q.GetSupply(ctx).GetTotal()

// 	return &types.QueryTotalSupplyResponse{Supply: totalSupply}, nil
// }

// // SupplyOf implements the Query/SupplyOf gRPC method
// func (q Keeper) SupplyOf(c context.Context, req *types.QuerySupplyOfRequest) (*types.QuerySupplyOfResponse, error) {
// 	if req == nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "empty request")
// 	}

// 	if req.Denom == "" {
// 		return nil, status.Errorf(codes.InvalidArgument, "invalid denom")
// 	}

// 	ctx := sdk.UnwrapSDKContext(c)
// 	supply := q.GetSupply(ctx).GetTotal().AmountOf(req.Denom)

// 	return &types.QuerySupplyOfResponse{Amount: supply}, nil
// }
