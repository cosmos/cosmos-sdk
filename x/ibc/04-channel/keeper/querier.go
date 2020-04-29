package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// QuerierChannels defines the sdk.Querier to query all the channels.
func QuerierChannels(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAllChannelsParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	channels := k.GetAllChannels(ctx)

	start, end := client.Paginate(len(channels), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		channels = []types.IdentifiedChannel{}
	} else {
		channels = channels[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, channels)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// QuerierConnectionChannels defines the sdk.Querier to query all the channels for a connection.
func QuerierConnectionChannels(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryConnectionChannelsParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	channels := k.GetAllChannels(ctx)

	connectionChannels := []types.IdentifiedChannel{}
	for _, channel := range channels {
		if channel.ConnectionHops[0] == params.Connection {
			connectionChannels = append(connectionChannels, channel)
		}
	}

	start, end := client.Paginate(len(connectionChannels), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		connectionChannels = []types.IdentifiedChannel{}
	} else {
		connectionChannels = channels[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, connectionChannels)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
