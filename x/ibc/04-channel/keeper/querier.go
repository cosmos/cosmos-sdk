package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// NewQuerier creates a querier for the IBC channel
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryChannel:
			return queryClientState(ctx, req, k)

		default:
			return nil, sdk.ErrUnknownRequest("unknown IBC channel query endpoint")
		}
	}
}

func queryClientState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryChannelParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	channel, found := k.GetChannel(ctx, params.PortID, params.ChannelID)
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, params.ChannelID)
	}

	bz, err := types.SubModuleCdc.MarshalJSON(channel)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
