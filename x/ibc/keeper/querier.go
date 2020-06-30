package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientkeeper "github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channelkeeper "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/keeper"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// NewQuerier creates a querier for the IBC module
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			res []byte
			err error
		)

		switch path[0] {
		case clienttypes.SubModuleName:
			switch path[1] {
			case clienttypes.QueryAllClients:
				res, err = clientkeeper.QuerierClients(ctx, req, k.ClientKeeper)
			default:
				err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", clienttypes.SubModuleName)
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
		case channeltypes.SubModuleName:
			switch path[1] {
			case channeltypes.QueryAllChannels:
				res, err = channelkeeper.QuerierChannels(ctx, req, k.ChannelKeeper)
			case channeltypes.QueryConnectionChannels:
				res, err = channelkeeper.QuerierConnectionChannels(ctx, req, k.ChannelKeeper)
			case channeltypes.QueryChannelClientState:
				res, err = channelkeeper.QuerierChannelClientState(ctx, req, k.ChannelKeeper)
			case channeltypes.QueryPacketCommitments:
				res, err = channelkeeper.QuerierPacketCommitments(ctx, req, k.ChannelKeeper)
			case channeltypes.QueryUnrelayedAcknowledgements:
				res, err = channelkeeper.QuerierUnrelayedAcknowledgements(ctx, req, k.ChannelKeeper)
			case channeltypes.QueryUnrelayedPacketSends:
				res, err = channelkeeper.QuerierUnrelayedPacketSends(ctx, req, k.ChannelKeeper)
			default:
				err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", channeltypes.SubModuleName)
			}
		default:
			err = sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown IBC query endpoint")
		}

		return res, err
	}
}
