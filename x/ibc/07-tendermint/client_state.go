package tendermint

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ clientexported.ClientState = ClientState{}

// ClientState from Tendermint tracks the current validator set, latest height,
// and a possible frozen height.
type ClientState struct {
	// Client ID
	ID string `json:"id" yaml:"id"`
	// Latest block height
	LatestHeight uint64 `json:"latest_height" yaml:"latest_height"`
	// Block height when the client was frozen due to a misbehaviour
	FrozenHeight uint64 `json:"frozen_height" yaml:"frozen_height"`
}

// NewClientState creates a new ClientState instance
func NewClientState(id string, latestHeight uint64) ClientState {
	return ClientState{
		ID:           id,
		LatestHeight: latestHeight,
		FrozenHeight: 0,
	}
}

// GetID returns the tendermint client state identifier.
func (cs ClientState) GetID() string {
	return cs.ID
}

// ClientType is tendermint.
func (cs ClientState) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// GetLatestHeight returns latest block height.
func (cs ClientState) GetLatestHeight() uint64 {
	return cs.LatestHeight
}

// IsFrozen returns true if the frozen height has been set.
func (cs ClientState) IsFrozen() bool {
	return cs.FrozenHeight != 0
}

func (cs ClientState) VerifyClientConsensusState(
	cdc *codec.Codec,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.ConsensusStatePath(cs.GetID(), height))
	if err != nil {
		return err
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienttypes.ErrClientFrozen
	}

	if proof == nil {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "proof cannot be empty")
	}

	bz, err := cdc.MarshalBinaryBare(consensusState)
	if err != nil {
		return err
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, bz); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedClientConsensusStateVerification, err.Error())
	}

	return nil
}

func (cs ClientState) VerifyConnectionState(
	cdc *codec.Codec,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienttypes.ErrClientFrozen
	}

	if proof == nil {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "proof cannot be empty")
	}

	bz, err := cdc.MarshalBinaryBare(connectionEnd)
	if err != nil {
		return err
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, bz); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedConnectionStateVerification, err.Error())
	}

	return nil
}

func (cs ClientState) VerifyChannelState(
	cdc *codec.Codec,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	channel channelexported.ChannelI,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienttypes.ErrClientFrozen
	}

	if proof == nil {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "proof cannot be empty")
	}

	bz, err := cdc.MarshalBinaryBare(channel)
	if err != nil {
		return err
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, bz); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedChannelStateVerification, err.Error())
	}

	return nil
}

func (cs ClientState) VerifyPacketCommitment(
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.PacketCommitmentPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienttypes.ErrClientFrozen
	}

	if proof == nil {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "proof cannot be empty")
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, commitmentBytes); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketCommitmentVerification, err.Error())
	}

	return nil
}

func (cs ClientState) VerifyPacketAcknowledgement(
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienttypes.ErrClientFrozen
	}

	if proof == nil {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "proof cannot be empty")
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, acknowledgement); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckVerification, err.Error())
	}

	return nil
}

func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienttypes.ErrClientFrozen
	}

	if proof == nil {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "proof cannot be empty")
	}

	if err := proof.VerifyNonMembership(consensusState.GetRoot(), path); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckAbsenceVerification, err.Error())
	}

	return nil
}

func (cs ClientState) VerifyNextSequenceRecv(
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	nextSequenceRecv uint64,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienttypes.ErrClientFrozen
	}

	if proof == nil {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "proof cannot be empty")
	}

	bz := sdk.Uint64ToBigEndian(nextSequenceRecv)

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, bz); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedNextSeqRecvVerification, err.Error())
	}

	return nil
}
