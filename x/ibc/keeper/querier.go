package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientkeeper "github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelkeeper "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/keeper"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// NewQuerier creates a querier for the IBC module
func NewQuerier(k Keeper, legacyQuerierCdc codec.JSONMarshaler) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			res []byte
			err error
		)

		switch path[0] {
		case clienttypes.SubModuleName:
			switch path[1] {
			case clienttypes.QueryAllClients:
				res, err = clientkeeper.QuerierClients(ctx, req, k.ClientKeeper, legacyQuerierCdc)
			default:
				err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", clienttypes.SubModuleName)
			}
		case connectiontypes.SubModuleName:
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", connectiontypes.SubModuleName)
		case channeltypes.SubModuleName:
			switch path[1] {
			case channeltypes.QueryChannelClientState:
				res, err = channelkeeper.QuerierChannelClientState(ctx, req, k.ChannelKeeper, legacyQuerierCdc)
			default:
				err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", channeltypes.SubModuleName)
			}
		default:
			err = sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown IBC query endpoint")
		}

		return res, err
	}
}
