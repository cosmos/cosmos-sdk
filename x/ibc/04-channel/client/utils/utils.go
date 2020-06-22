package utils

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// QueryPacket returns a packet from the store
func QueryPacket(
	clientCtx client.Context, portID, channelID string,
	sequence, timeoutHeight, timeoutTimestamp uint64, prove bool,
) (*types.QueryPacketResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyPacketCommitment(portID, channelID, sequence),
		Prove: prove,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	channelRes, err := QueryChannel(clientCtx, portID, channelID, prove)
	if err != nil {
		return nil, err
	}

	proofBz, err := clientCtx.Codec.MarshalBinaryBare(res.Proof)
	if err != nil {
		return nil, err
	}

	destPortID := channelRes.Channel.Counterparty.PortID
	destChannelID := channelRes.Channel.Counterparty.ChannelID

	packet := types.NewPacket(
		res.Value,
		sequence,
		portID,
		channelID,
		destPortID,
		destChannelID,
		timeoutHeight,
		timeoutTimestamp,
	)

	// FIXME: res.Height+1 is hack, fix later
	return types.NewQueryPacketResponse(portID, channelID, sequence, packet, proofBz, res.Height+1), nil
}

// QueryChannel queries the store to get a channel and a merkle proof.
func QueryChannel(
	clientCtx client.Context, portID, channelID string, prove bool,
) (*types.QueryChannelResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyChannel(portID, channelID),
		Prove: prove,
	}

	res, err := clientCtx.QueryABCI(req)
	if res.Value == nil || err != nil {
		return nil, err
	}

	var channel types.Channel
	if err := clientCtx.Codec.UnmarshalBinaryBare(res.Value, &channel); err != nil {
		return nil, err
	}

	proofBz, err := clientCtx.Codec.MarshalBinaryBare(res.Proof)
	if err != nil {
		return nil, err
	}

	return types.NewQueryChannelResponse(portID, channelID, channel, proofBz, res.Height), nil
}

// QueryChannelClientState uses the channel Querier to return the ClientState of
// a Channel.
func QueryChannelClientState(clientCtx client.Context, portID, channelID string) (clientexported.ClientState, int64, error) {
	params := types.NewQueryChannelClientStateRequest(portID, channelID)
	bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryChannelClientState)
	res, height, err := clientCtx.QueryWithData(route, bz)
	if err != nil {
		return nil, 0, err
	}

	var clientState clientexported.ClientState
	err = clientCtx.JSONMarshaler.UnmarshalJSON(res, &clientState)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal connections: %w", err)
	}
	return clientState, height, nil
}
