package utils

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// QueryPacket returns a packet from the store
func QueryPacket(
	ctx context.CLIContext, portID, channelID string,
	sequence, timeout uint64, prove bool,
) (types.PacketResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyPacketCommitment(portID, channelID, sequence),
		Prove: prove,
	}

	res, err := ctx.QueryABCI(req)
	if err != nil {
		return types.PacketResponse{}, err
	}

	channelRes, err := QueryChannel(ctx, portID, channelID, prove)
	if err != nil {
		return types.PacketResponse{}, err
	}

	destPortID := channelRes.Channel.Counterparty.PortID
	destChannelID := channelRes.Channel.Counterparty.ChannelID

	var data exported.PacketDataI
	// TODO: commitment data is stored, not the data
	// but we are unmarshalling the commitment in a json format
	// because the current ICS 20 implementation uses json commitment form
	// need to be changed to use external source of packet(e.g. event logs)
	err = ctx.Codec.UnmarshalJSON(res.Value, &data)
	if err != nil {
		return types.PacketResponse{}, err
	}

	packet := types.NewPacket(
		data,
		sequence,
		portID,
		channelID,
		destPortID,
		destChannelID,
	)

	// FIXME: res.Height+1 is hack, fix later
	return types.NewPacketResponse(portID, channelID, sequence, packet, res.Proof, res.Height+1), nil
}

// QueryChannel queries the store to get a channel and a merkle proof.
func QueryChannel(
	ctx context.CLIContext, portID, channelID string, prove bool,
) (types.ChannelResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyChannel(portID, channelID),
		Prove: prove,
	}

	res, err := ctx.QueryABCI(req)
	if res.Value == nil || err != nil {
		return types.ChannelResponse{}, err
	}

	var channel types.Channel
	if err := ctx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &channel); err != nil {
		return types.ChannelResponse{}, err
	}
	return types.NewChannelResponse(portID, channelID, channel, res.Proof, res.Height), nil
}
