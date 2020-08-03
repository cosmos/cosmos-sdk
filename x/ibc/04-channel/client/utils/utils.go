package utils

import (
	"context"
	"encoding/binary"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// QueryPacketCommitment returns a packet commitment.
// If prove is true, it performs an ABCI store query in order to retrieve the merkle proof. Otherwise,
// it uses the gRPC query client.
func QueryPacketCommitment(
	clientCtx client.Context, portID, channelID string,
	sequence uint64, prove bool,
) (*types.QueryPacketCommitmentResponse, error) {
	if prove {
		return queryPacketCommitmentABCI(clientCtx, portID, channelID, sequence)
	}

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryPacketCommitmentRequest{
		PortID:    portID,
		ChannelID: channelID,
		Sequence:  sequence,
	}

	return queryClient.PacketCommitment(context.Background(), req)
}

func queryPacketCommitmentABCI(
	clientCtx client.Context, portID, channelID string, sequence uint64,
) (*types.QueryPacketCommitmentResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyPacketCommitment(portID, channelID, sequence),
		Prove: true,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	proofBz, err := clientCtx.Codec.MarshalBinaryBare(res.ProofOps)
	if err != nil {
		return nil, err
	}

	// FIXME: res.Height+1 is hack, fix later
	return types.NewQueryPacketCommitmentResponse(portID, channelID, sequence, res.Value, proofBz, res.Height+1), nil
}

// QueryChannel returns a channel end.
// If prove is true, it performs an ABCI store query in order to retrieve the merkle proof. Otherwise,
// it uses the gRPC query client.
func QueryChannel(
	clientCtx client.Context, portID, channelID string, prove bool,
) (*types.QueryChannelResponse, error) {
	if prove {
		return queryChannelABCI(clientCtx, portID, channelID)
	}

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryChannelRequest{
		PortID:    portID,
		ChannelID: channelID,
	}

	return queryClient.Channel(context.Background(), req)
}

func queryChannelABCI(clientCtx client.Context, portID, channelID string) (*types.QueryChannelResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyChannel(portID, channelID),
		Prove: true,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	var channel types.Channel
	if err := clientCtx.Codec.UnmarshalBinaryBare(res.Value, &channel); err != nil {
		return nil, err
	}

	proofBz, err := clientCtx.Codec.MarshalBinaryBare(res.ProofOps)
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
		return nil, 0, fmt.Errorf("failed to unmarshal client state: %w", err)
	}
	return clientState, height, nil
}

// QueryChannelConsensusState uses the channel Querier to return the ConsensusState
// of a Channel.
func QueryChannelConsensusState(clientCtx client.Context, portID, channelID string) (clientexported.ConsensusState, int64, error) {
	params := types.NewQueryChannelConsensusStateRequest(portID, channelID)
	bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryChannelConsensusState)
	res, height, err := clientCtx.QueryWithData(route, bz)
	if err != nil {
		return nil, 0, err
	}

	var consensusState clientexported.ConsensusState
	err = clientCtx.JSONMarshaler.UnmarshalJSON(res, &consensusState)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal consensus state: %w", err)
	}
	return consensusState, height, nil
}

// QueryCounterpartyConsensusState uses the channel Querier to return the
// counterparty ConsensusState given the source port ID and source channel ID.
func QueryCounterpartyConsensusState(clientCtx client.Context, portID, channelID string) (clientexported.ConsensusState, int64, error) {
	channelRes, err := QueryChannel(clientCtx, portID, channelID, false)
	if err != nil {
		return nil, 0, err
	}

	counterparty := channelRes.Channel.Counterparty
	clientState, height, err := QueryChannelConsensusState(clientCtx, counterparty.PortID, counterparty.ChannelID)
	if err != nil {
		return nil, 0, err
	}

	return clientState, height, nil
}

// QueryNextSequenceReceive returns the next sequence receive.
// If prove is true, it performs an ABCI store query in order to retrieve the merkle proof. Otherwise,
// it uses the gRPC query client.
func QueryNextSequenceReceive(
	clientCtx client.Context, portID, channelID string, prove bool,
) (*types.QueryNextSequenceReceiveResponse, error) {
	if prove {
		return queryNextSequenceRecvABCI(clientCtx, portID, channelID)
	}

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryNextSequenceReceiveRequest{
		PortID:    portID,
		ChannelID: channelID,
	}

	return queryClient.NextSequenceReceive(context.Background(), req)
}

func queryNextSequenceRecvABCI(clientCtx client.Context, portID, channelID string) (*types.QueryNextSequenceReceiveResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyNextSequenceRecv(portID, channelID),
		Prove: true,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	proofBz, err := clientCtx.Codec.MarshalBinaryBare(res.ProofOps)
	if err != nil {
		return nil, err
	}

	sequence := binary.BigEndian.Uint64(res.Value)
	return types.NewQueryNextSequenceReceiveResponse(portID, channelID, sequence, proofBz, res.Height), nil
}
