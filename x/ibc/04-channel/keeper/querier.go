package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// QuerierChannelClientState defines the sdk.Querier to query all the ClientState
// associated with a given Channel.
func QuerierChannelClientState(ctx sdk.Context, abciReq abci.RequestQuery, k Keeper) ([]byte, error) {
	var req types.QueryChannelClientStateRequest

	if err := k.cdc.UnmarshalJSON(abciReq.Data, &req); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	channel, found := k.GetChannel(ctx, req.PortID, req.ChannelID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrChannelNotFound, req.PortID, req.ChannelID)
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, sdkerrors.Wrapf(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return nil, sdkerrors.Wrapf(clienttypes.ErrClientNotFound, connection.ClientID)
	}

	res, err := codec.MarshalJSONIndent(k.cdc, clientState)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
