package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// QuerierChannel defines the sdk.Querier to query a module's channel
func QuerierChannel(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryChannelParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	channel, found := k.GetChannel(ctx, params.PortID, params.ChannelID)
	if !found {
		return nil, sdk.ConvertError(types.ErrChannelNotFound(k.codespace, params.PortID, params.ChannelID))
	}

	bz, err := types.SubModuleCdc.MarshalJSON(channel)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// QuerierChannels defines the sdk.Querier to query all the light client states.
func QuerierChannels(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryAllChannelsParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to unmarshall json: %s", err.Error()))
	}

	clients := k.GetAllChannels(ctx)

	start, end := client.Paginate(len(clients), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		clients = []types.State{}
	} else {
		clients = clients[start:end]
	}

	res, err := types.SubModuleCdc.MarshalJSON(clients)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to unmarshall json: %s", err.Error()))
	}

	return res, nil
}
