package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// QuerierConnections defines the sdk.Querier to query all the connections.
func QuerierConnections(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAllConnectionsParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	connections := k.GetAllConnections(ctx)

	start, end := client.Paginate(len(connections), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		connections = []types.ConnectionEnd{}
	} else {
		connections = connections[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, connections)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// QuerierClientConnections defines the sdk.Querier to query the client connections
func QuerierClientConnections(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryClientConnectionsParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	clientConnectionPaths, found := k.GetClientConnectionPaths(ctx, params.ClientID)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrClientConnectionPathsNotFound, params.ClientID)
	}

	bz, err := types.SubModuleCdc.MarshalJSON(clientConnectionPaths)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// QuerierAllClientConnections defines the sdk.Querier to query the connections paths for clients.
func QuerierAllClientConnections(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAllConnectionsParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	clientsConnectionPaths := k.GetAllClientConnectionPaths(ctx)

	start, end := client.Paginate(len(clientsConnectionPaths), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		clientsConnectionPaths = []types.ConnectionPaths{}
	} else {
		clientsConnectionPaths = clientsConnectionPaths[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, clientsConnectionPaths)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
