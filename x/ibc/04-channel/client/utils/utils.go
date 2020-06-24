package utils

import (
	"encoding/binary"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// QueryPacketCommitment returns a packet commitment from the store
func QueryPacketCommitment(
	clientCtx client.Context, portID, channelID string,
	sequence, timeoutHeight, timeoutTimestamp uint64, prove bool,
) (*types.QueryPacketCommitmentResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyPacketCommitment(portID, channelID, sequence),
		Prove: prove,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	proofBz, err := clientCtx.Codec.MarshalBinaryBare(res.Proof)
	if err != nil {
		return nil, err
	}

	// FIXME: res.Height+1 is hack, fix later
	return types.NewQueryPacketCommitmentResponse(portID, channelID, sequence, res.Value, proofBz, res.Height+1), nil
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
	if err != nil {
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

// QueryNextSequenceReceive queries the store to get the next receive sequence and
// a merkle proof.
func QueryNextSequenceReceive(
	clientCtx client.Context, portID, channelID string, prove bool,
) (*types.QueryNextSequenceReceiveResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyNextSequenceRecv(portID, channelID),
		Prove: prove,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	sequence := binary.BigEndian.Uint64(res.Value)
	return types.NewQueryNextSequenceReceiveResponse(portID, channelID, sequence, res.Proof, res.Height), nil
}
