package utils

import (
	"context"
	"encoding/binary"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientutils "github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
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
		PortId:    portID,
		ChannelId: channelID,
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

	cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

	proofBz, err := cdc.MarshalBinaryBare(res.ProofOps)
	if err != nil {
		return nil, err
	}

	// FIXME: height + 1 is returned as the proof height
	// Issue: https://github.com/cosmos/cosmos-sdk/issues/6567
	// TODO: retrieve epoch number from chain-id
	height := clienttypes.NewHeight(0, uint64(res.Height+1))
	return types.NewQueryPacketCommitmentResponse(portID, channelID, sequence, res.Value, proofBz, height), nil
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
		PortId:    portID,
		ChannelId: channelID,
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

	cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

	var channel types.Channel
	if err := cdc.UnmarshalBinaryBare(res.Value, &channel); err != nil {
		return nil, err
	}

	proofBz, err := cdc.MarshalBinaryBare(res.ProofOps)
	if err != nil {
		return nil, err
	}

	// TODO: retrieve epoch number from chain-id
	height := clienttypes.NewHeight(0, uint64(res.Height))
	return types.NewQueryChannelResponse(portID, channelID, channel, proofBz, height), nil
}

// QueryChannelClientState returns the ClientState of a channel end. If
// prove is true, it performs an ABCI store query in order to retrieve the
// merkle proof. Otherwise, it uses the gRPC query client.
func QueryChannelClientState(
	clientCtx client.Context, portID, channelID string, prove bool,
) (*types.QueryChannelClientStateResponse, error) {

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryChannelClientStateRequest{
		PortId:    portID,
		ChannelId: channelID,
	}

	res, err := queryClient.ChannelClientState(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if prove {
		clientStateRes, err := clientutils.QueryClientStateABCI(clientCtx, res.IdentifiedClientState.ClientId)
		if err != nil {
			return nil, err
		}

		// use client state returned from ABCI query in case query height differs
		identifiedClientState := clienttypes.IdentifiedClientState{
			ClientId:    res.IdentifiedClientState.ClientId,
			ClientState: clientStateRes.ClientState,
		}
		res = types.NewQueryChannelClientStateResponse(identifiedClientState, clientStateRes.Proof, clientStateRes.ProofHeight)
	}

	return res, nil
}

// QueryChannelConsensusState returns the ConsensusState of a channel end. If
// prove is true, it performs an ABCI store query in order to retrieve the
// merkle proof. Otherwise, it uses the gRPC query client.
func QueryChannelConsensusState(
	clientCtx client.Context, portID, channelID string, height clienttypes.Height, prove bool,
) (*types.QueryChannelConsensusStateResponse, error) {

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryChannelConsensusStateRequest{
		PortId:      portID,
		ChannelId:   channelID,
		EpochNumber: height.EpochNumber,
		EpochHeight: height.EpochHeight,
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
		consensusStateRes, err := clientutils.QueryConsensusStateABCI(clientCtx, res.ClientId, consensusState.GetHeight())
		if err != nil {
			return nil, err
		}

		res = types.NewQueryChannelConsensusStateResponse(res.ClientId, consensusStateRes.ConsensusState, height, consensusStateRes.Proof, consensusStateRes.ProofHeight)
	}

	return res, nil
}

// QueryLatestConsensusState uses the channel Querier to return the
// latest ConsensusState given the source port ID and source channel ID.
func QueryLatestConsensusState(
	clientCtx client.Context, portID, channelID string,
) (exported.ConsensusState, clienttypes.Height, error) {
	clientRes, err := QueryChannelClientState(clientCtx, portID, channelID, false)
	if err != nil {
		return nil, clienttypes.Height{}, err
	}
	clientState, err := clienttypes.UnpackClientState(clientRes.IdentifiedClientState.ClientState)
	if err != nil {
		return nil, clienttypes.Height{}, err
	}

	clientHeight, ok := clientState.GetLatestHeight().(clienttypes.Height)
	if !ok {
		return nil, clienttypes.Height{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "invalid height type. expected type: %T, got: %T",
			clienttypes.Height{}, clientHeight)
	}
	res, err := QueryChannelConsensusState(clientCtx, portID, channelID, clientHeight, false)
	if err != nil {
		return nil, clienttypes.Height{}, err
	}

	consensusState, err := clienttypes.UnpackConsensusState(res.ConsensusState)
	if err != nil {
		return nil, clienttypes.Height{}, err
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
		PortId:    portID,
		ChannelId: channelID,
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

	cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

	proofBz, err := cdc.MarshalBinaryBare(res.ProofOps)
	if err != nil {
		return nil, err
	}

	sequence := binary.BigEndian.Uint64(res.Value)
	// TODO: retrieve epoch number from chain-id
	height := clienttypes.NewHeight(0, uint64(res.Height))
	return types.NewQueryNextSequenceReceiveResponse(portID, channelID, sequence, proofBz, height), nil
}
