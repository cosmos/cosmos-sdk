package utils

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// QueryPacket returns a packet from the store
func QueryPacket(
	ctx client.CLIContext, portID, channelID string,
	sequence, timeout uint64, queryRoute string,
) (types.PacketResponse, error) {
	var packetRes types.PacketResponse

	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyPacketCommitment(portID, channelID, sequence),
		Prove: true,
	}

	res, err := ctx.QueryABCI(req)
	if err != nil {
		return packetRes, err
	}

	channel, err := QueryChannel(ctx, portID, channelID, queryRoute)
	if err != nil {
		return packetRes, err
	}

	destPortID := channel.Channel.Counterparty.PortID
	destChannelID := channel.Channel.Counterparty.ChannelID

	packet := types.NewFullPacket(
		sequence,
		timeout,
		portID,
		channelID,
		destPortID,
		destChannelID,
		res.Value,
	)

	// FIXME: res.Height+1 is hack, fix later
	return types.NewPacketResponse(packet, res.Proof, res.Height+1), nil
}

// QueryChannel returns a channel from the store
func QueryChannel(ctx client.CLIContext, portID string, channelID string, queryRoute string) (types.ChannelResponse, error) {
	var connRes types.ChannelResponse

	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyChannel(portID, channelID),
		Prove: true,
	}

	res, err := ctx.QueryABCI(req)
	if res.Value == nil || err != nil {
		return connRes, err
	}

	var channel types.Channel
	if err := ctx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &channel); err != nil {
		return connRes, err
	}
	return types.NewChannelResponse(portID, channelID, channel, res.Proof, res.Height), nil
}
