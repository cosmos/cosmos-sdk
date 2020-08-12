package utils

import (
	"context"
	"encoding/binary"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	clientutils "github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
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

	proofBz, err := clientCtx.LegacyAmino.MarshalBinaryBare(res.Proof)
	if err != nil {
		return nil, err
	}

	// FIXME: height + 1 is returned as the proof height
	// Issue: https://github.com/cosmos/cosmos-sdk/issues/6567
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
	if err := clientCtx.LegacyAmino.UnmarshalBinaryBare(res.Value, &channel); err != nil {
		return nil, err
	}

	proofBz, err := clientCtx.LegacyAmino.MarshalBinaryBare(res.Proof)
	if err != nil {
		return nil, err
	}

	return types.NewQueryChannelResponse(portID, channelID, channel, proofBz, res.Height), nil
}

// QueryChannelClientState returns the ClientState of a channel end. If
// prove is true, it performs an ABCI store query in order to retrieve the
// merkle proof. Otherwise, it uses the gRPC query client.
func QueryChannelClientState(
	clientCtx client.Context, portID, channelID string, prove bool,
) (*types.QueryChannelClientStateResponse, error) {

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryChannelClientStateRequest{
		PortID:    portID,
		ChannelID: channelID,
	}

	res, err := queryClient.ChannelClientState(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if prove {
		clientState, proof, proofHeight, err := clientutils.QueryClientStateABCI(clientCtx, res.IdentifiedClientState.ID)
		if err != nil {
			return nil, err
		}

		// use client state returned from ABCI query in case query height differs
		identifiedClientState := clienttypes.NewIdentifiedClientState(res.IdentifiedClientState.ID, clientState)
		res = types.NewQueryChannelClientStateResponse(identifiedClientState, proof, int64(proofHeight))
	}

	return res, nil
}

// QueryChannelConsensusState returns the ConsensusState of a channel end. If
// prove is true, it performs an ABCI store query in order to retrieve the
// merkle proof. Otherwise, it uses the gRPC query client.
func QueryChannelConsensusState(
	clientCtx client.Context, portID, channelID string, prove bool,
) (*types.QueryChannelConsensusStateResponse, error) {

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryChannelConsensusStateRequest{
		PortID:    portID,
		ChannelID: channelID,
	}

	res, err := queryClient.ChannelConsensusState(context.Background(), req)
	if err != nil {
		return nil, err
	}

	consensusState, err := clienttypes.UnpackConsensusState(res.ConsensusState)
	if err != nil {
		return nil, err
	}

	if prove {
		consensusState, proof, proofHeight, err := clientutils.QueryConsensusStateABCI(clientCtx, res.ClientID, consensusState.GetHeight())
		if err != nil {
			return nil, err
		}

		// use consensus state returned from ABCI query in case query height differs
		anyConsensusState, err := clienttypes.PackConsensusState(consensusState)
		if err != nil {
			return nil, err
		}

		res = types.NewQueryChannelConsensusStateResponse(res.ClientID, anyConsensusState, consensusState.GetHeight(), proof, int64(proofHeight))
	}

	return res, nil
}

// QueryCounterpartyConsensusState uses the channel Querier to return the
// counterparty ConsensusState given the source port ID and source channel ID.
func QueryCounterpartyConsensusState(
	clientCtx client.Context, portID, channelID string,
) (clientexported.ConsensusState, uint64, error) {
	channelRes, err := QueryChannel(clientCtx, portID, channelID, false)
	if err != nil {
		return nil, 0, err
	}

	counterparty := channelRes.Channel.Counterparty
	res, err := QueryChannelConsensusState(clientCtx, counterparty.PortID, counterparty.ChannelID, false)
	if err != nil {
		return nil, 0, err
	}

	consensusState, err := clienttypes.UnpackConsensusState(res.ConsensusState)
	if err != nil {
		return nil, 0, err
	}

	return consensusState, res.ProofHeight, nil
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

	proofBz, err := clientCtx.LegacyAmino.MarshalBinaryBare(res.Proof)
	if err != nil {
		return nil, err
	}

	sequence := binary.BigEndian.Uint64(res.Value)
	return types.NewQueryNextSequenceReceiveResponse(portID, channelID, sequence, proofBz, res.Height), nil
}
