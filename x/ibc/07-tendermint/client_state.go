package tendermint

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienterrors "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
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

// GetSequence returns latest block height.
func (cs ClientState) GetSequence() uint64 {
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
	clientID string,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.ConsensusStatePath(clientID, height))
	if err != nil {
		return nil
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienterrors.ErrClientFrozen
	}

	bz, err := cdc.MarshalBinaryBare(consensusState)
	if err != nil {
		return err
	}

	if ok := proof.VerifyMembership(consensusState.GetRoot(), path, bz); !ok {
		return clienterrors.ErrFailedClientConsensusStateVerification
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
		return nil
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienterrors.ErrClientFrozen
	}

	bz, err := cdc.MarshalBinaryBare(connectionEnd)
	if err != nil {
		return err
	}

	if ok := proof.VerifyMembership(consensusState.GetRoot(), path, bz); !ok {
		return clienterrors.ErrFailedConnectionStateVerification
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
		return nil
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienterrors.ErrClientFrozen
	}

	bz, err := cdc.MarshalBinaryBare(channel)
	if err != nil {
		return err
	}

	if ok := proof.VerifyMembership(consensusState.GetRoot(), path, bz); !ok {
		return clienterrors.ErrFailedChannelStateVerification
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
		return nil
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienterrors.ErrClientFrozen
	}

	if ok := proof.VerifyMembership(consensusState.GetRoot(), path, commitmentBytes); !ok {
		return clienterrors.ErrFailedPacketCommitmentVerification
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
		return nil
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienterrors.ErrClientFrozen
	}

	if ok := proof.VerifyMembership(consensusState.GetRoot(), path, acknowledgement); !ok {
		return clienterrors.ErrFailedPacketAckVerification
	}

	return nil
}

func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	height uint64,
	prefix commitment.PrefixI, proof commitment.ProofI,
	portID, channelID string,
	sequence uint64, consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return nil
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienterrors.ErrClientFrozen
	}

	if ok := proof.VerifyNonMembership(consensusState.GetRoot(), path); !ok {
		return clienterrors.ErrFailedPacketAckAbsenceVerification
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
		return nil
	}

	if cs.LatestHeight < height {
		return ibctypes.ErrInvalidHeight
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienterrors.ErrClientFrozen
	}

	bz := sdk.Uint64ToBigEndian(nextSequenceRecv)

	if ok := proof.VerifyMembership(consensusState.GetRoot(), path, bz); !ok {
		return clienterrors.ErrFailedNextSeqRecvVerification
	}

	return nil
}
