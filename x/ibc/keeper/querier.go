package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

// NewQuerier creates a querier for the IBC module
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case client.SubModuleName:
			switch path[1] {
			case client.QueryClientState:
				return client.QuerierClientState(ctx, req, k.ClientKeeper)
			case client.QueryConsensusState:
				return client.QuerierConsensusState(ctx, req, k.ClientKeeper)
			case client.QueryVerifiedRoot:
				return client.QuerierVerifiedRoot(ctx, req, k.ClientKeeper)
			default:
				return nil, sdk.ErrUnknownRequest(fmt.Sprintf("unknown IBC %s query endpoint", client.SubModuleName))
			}
		case connection.SubModuleName:
			switch path[1] {
			case connection.QueryConnection:
				return connection.QuerierConnection(ctx, req, k.ConnectionKeeper)
			case connection.QueryClientConnections:
				return connection.QuerierClientConnections(ctx, req, k.ConnectionKeeper)
			default:
				return nil, sdk.ErrUnknownRequest(fmt.Sprintf("unknown IBC %s query endpoint", connection.SubModuleName))
			}
		case channel.SubModuleName:
			switch path[1] {
			case channel.QueryChannel:
				return channel.QuerierChannel(ctx, req, k.ChannelKeeper)
			default:
				return nil, sdk.ErrUnknownRequest(fmt.Sprintf("unknown IBC %s query endpoint", channel.SubModuleName))
			}
		default:
			return nil, sdk.ErrUnknownRequest("unknown IBC query endpoint")
		}
	}
}
