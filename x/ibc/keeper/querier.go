package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

// NewQuerier creates a querier for the IBC module
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			res []byte
			err error
		)

		switch path[0] {
		case client.SubModuleName:
			switch path[1] {
			case client.QueryAllClients:
				res, err = client.QuerierClients(ctx, req, k.ClientKeeper)
			default:
				err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", client.SubModuleName)
			}
		case connection.SubModuleName:
			switch path[1] {
			case connection.QueryAllConnections:
				res, err = connection.QuerierConnections(ctx, req, k.ConnectionKeeper)
			case connection.QueryAllClientConnections:
				res, err = connection.QuerierAllClientConnections(ctx, req, k.ConnectionKeeper)
			case connection.QueryClientConnections:
				res, err = connection.QuerierClientConnections(ctx, req, k.ConnectionKeeper)
			default:
				err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", connection.SubModuleName)
			}
		case channel.SubModuleName:
			switch path[1] {
			case channel.QueryAllChannels:
				res, err = channel.QuerierChannels(ctx, req, k.ChannelKeeper)
			case channel.QueryConnectionChannels:
				res, err = channel.QuerierConnectionChannels(ctx, req, k.ChannelKeeper)
			case channel.QueryPacketCommitments:
				res, err = channel.QuerierPacketCommitments(ctx, req, k.ChannelKeeper)
			case channel.QueryUnrelayedAcknowledgements:
				res, err = channel.QuerierUnrelayedAcknowledgements(ctx, req, k.ChannelKeeper)
			case channel.QueryUnrelayedPacketSends:
				res, err = channel.QuerierUnrelayedPacketSends(ctx, req, k.ChannelKeeper)
			default:
				err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", channel.SubModuleName)
			}
		default:
			err = sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown IBC query endpoint")
		}

		return res, err
	}
}
