package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// QuerierConnection defines the sdk.Querier to query a connection end
func QuerierConnection(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryConnectionParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	connection, found := k.GetConnection(ctx, params.ConnectionID)
	if !found {
		return nil, sdk.ConvertError(types.ErrConnectionNotFound(k.codespace, params.ConnectionID))
	}

	bz, err := types.SubModuleCdc.MarshalJSON(connection)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// QuerierClientConnections defines the sdk.Querier to query the client connections
func QuerierClientConnections(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryClientConnectionsParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	clientConnectionPaths, found := k.GetClientConnectionPaths(ctx, params.ClientID)
	if !found {
		return nil, sdk.ConvertError(types.ErrClientConnectionPathsNotFound(k.codespace, params.ClientID))
	}

	bz, err := types.SubModuleCdc.MarshalJSON(clientConnectionPaths)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// QueryAllConnections defines the sdk.Querier to query all the connections.
func QueryAllConnections(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAllConnectionsParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
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
