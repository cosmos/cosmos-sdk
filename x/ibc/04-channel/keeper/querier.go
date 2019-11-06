package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

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
		return nil, types.ErrChannelNotFound(k.codespace, params.PortID, params.ChannelID)
	}

	bz, err := types.SubModuleCdc.MarshalJSON(channel)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
