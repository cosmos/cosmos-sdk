package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/codec"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	sdkerrors "github.com/KiraCore/cosmos-sdk/types/errors"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/types"
)

// QuerierClients defines the sdk.Querier to query all the light client states.
func QuerierClients(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc codec.JSONMarshaler) ([]byte, error) {
	var params types.QueryAllClientsParams

	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	clients := k.GetAllClients(ctx)

	start, end := client.Paginate(len(clients), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		clients = []exported.ClientState{}
	} else {
		clients = clients[start:end]
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, clients)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
