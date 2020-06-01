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

// QuerierPacketCommitments defines the sdk.Querier to query all packet commitments on a
// specified channel.
func QuerierPacketCommitments(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryPacketCommitmentsParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	packetCommitments := k.GetAllPacketCommitmentsAtChannel(ctx, params.PortID, params.ChannelID)
	sequences := make([]uint64, 0, len(packetCommitments))

	for _, pc := range packetCommitments {
		sequences = append(sequences, pc.Sequence)
	}

	start, end := client.Paginate(len(sequences), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		sequences = []uint64{}
	} else {
		sequences = sequences[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, sequences)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// QuerierUnrelayedAcknowledgements defines the sdk.Querier to query all unrelayed
// acknowledgements for a specified channel end.
func QuerierUnrelayedAcknowledgements(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	return queryUnrelayedPackets(ctx, req, k, true)
}

// QuerierUnrelayedPacketSends defines the sdk.Querier to query all unrelayed packets for a
// specified channel end.
func QuerierUnrelayedPacketSends(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	return queryUnrelayedPackets(ctx, req, k, false)
}

// helper function to query for unrelayed packets as specified by the isForAcks boolean. If
// set to true it will return unrelayed acknowledgements otherwise it will return unrelayed
// packet sends.
func queryUnrelayedPackets(ctx sdk.Context, req abci.RequestQuery, k Keeper, isForAcks bool) ([]byte, error) {
	var params types.QueryUnrelayedPacketsParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	var unrelayedPackets []uint64
	for _, seq := range params.Sequences {
		if _, found := k.GetPacketAcknowledgement(ctx, params.PortID, params.ChannelID, seq); found == isForAcks {
			unrelayedPackets = append(unrelayedPackets, seq)
		}
	}

	start, end := client.Paginate(len(unrelayedPackets), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		unrelayedPackets = []uint64{}
	} else {
		unrelayedPackets = unrelayedPackets[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, unrelayedPackets)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil

}
