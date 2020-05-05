package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ clientexported.ClientState = ClientState{}

// ClientState of a Solo Machine represents whether or not the client is frozen.
type ClientState struct {
	// Client ID
	ID string `json:"id" yaml:"id"`

	// Frozen status of the client
	Frozen bool `json:"frozen" yaml:"frozen"`

	// Current consensus state of the client
	ConsensusState ConsensusState `json:"consensus_state" yaml:"consensus_state"`
}

// InitializeFromMsg creates a solo machine client from a MsgCreateClient
func InitializeFromMsg(msg MsgCreateClient) (ClientState, error) {
	return Initialize(msg.GetClientID(), msg.ConsensusState)
}

// Initialize creates an unfrozen client with the initial consensus state
func Initialize(id string, consensusState ConsensusState) (ClientState, error) {
	return ClientState{
		ID:             id,
		Frozen:         false,
		ConsensusState: consensusState,
	}, nil
}

// GetID returns the solo machine client state identifier.
func (cs ClientState) GetID() string {
	return cs.ID
}

// GetChainID returns an empty string.
func (cs ClientState) GetChainID() string {
	return ""
}

// ClientType is Solo Machine.
func (cs ClientState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetLatestHeight returns the latest sequence number.
func (cs ClientState) GetLatestHeight() uint64 {
	return cs.ConsensusState.Sequence
}

// IsFrozen returns true if the client is frozen
func (cs ClientState) IsFrozen() bool {
	return cs.Frozen
}

// VerifyClientConsensusState verifies a proof of the consensus state of the
// Solo Machine client stored on the target machine.
func (cs ClientState) VerifyClientConsensusState(
	cdc *codec.Codec,
	root commitmentexported.Root,
	height uint64,
	counterpartyClientIdentifier string,
	consensusHeight uint64,
	prefix commitmentexported.Prefix,
	proof commitmentexported.Proof,
	consensusState clientexported.ConsensusState,
) error {
	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + ibctypes.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	if cs.IsFrozen() {
		return clienttypes.ErrClientFrozen
	}

	bz, err := cdc.MarshalBinaryBare(consensusState)
	if err != nil {
		return err
	}

	// value = sequence + path + consensus state
	value := append(
		combineSequenceAndPath(cs.ConsensusState.Sequence, path),
		bz...,
	)
	if cs.ConsensusState.PublicKey.VerifyBytes(value, proof) {
		return sdkerrors.Wrap(clienttypes.ErrFailedClientConsensusStateVerification, "failed to verify proof against current public key, sequence, and consensus state")
	}

	//cs.ConsensusState.Sequence += 1
	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (cs ClientState) VerifyConnectionState(
	cdc *codec.Codec,
	height uint64,
	prefix commitmentexported.Prefix,
	proof commitmentexported.Proof,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	if cs.IsFrozen() {
		return clienttypes.ErrClientFrozen
	}

	bz, err := cdc.MarshalBinaryBare(connectionEnd)
	if err != nil {
		return err
	}

	// value = sequence + path + connection end
	value := append(
		combineSequenceAndPath(cs.ConsensusState.Sequence, path),
		bz...,
	)
	if cs.ConsensusState.PublicKey.VerifyBytes(value, proof) {
		return sdkerrors.Wrap(
			clienttypes.ErrFailedConnectionStateVerification,
			"failed to verify proof against current public key, sequence, and connection state",
		)
	}

	//cs.ConsensusState.Sequence += 1
	return nil

}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (cs ClientState) VerifyChannelState(
	cdc *codec.Codec,
	height uint64,
	prefix commitmentexported.Prefix,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	channel channelexported.ChannelI,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	if cs.IsFrozen() {
		return clienttypes.ErrClientFrozen
	}

	bz, err := cdc.MarshalBinaryBare(channel)
	if err != nil {
		return err
	}

	// value = sequence + path + channel
	value := append(
		combineSequenceAndPath(cs.ConsensusState.Sequence, path),
		bz...,
	)
	if cs.ConsensusState.PublicKey.VerifyBytes(value, proof) {
		return sdkerrors.Wrap(
			clienttypes.ErrFailedChannelStateVerification,
			"failed to verify proof against current public key, sequence, and channel state",
		)
	}

	//cs.ConsensusState.Sequence += 1
	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	height uint64,
	prefix commitmentexported.Prefix,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.PacketCommitmentPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if cs.IsFrozen() {
		return clienttypes.ErrClientFrozen
	}

	// value = sequence + path + commitment bytes
	value := append(
		combineSequenceAndPath(cs.ConsensusState.Sequence, path),
		commitmentBytes...,
	)
	if cs.ConsensusState.PublicKey.VerifyBytes(value, proof) {
		return sdkerrors.Wrap(
			clienttypes.ErrFailedPacketCommitmentVerification,
			"failed to verify proof against current public key, sequence, and packet commitment",
		)
	}

	//cs.ConsensusState.Sequence += 1
	return nil

}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	height uint64,
	prefix commitmentexported.Prefix,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if cs.IsFrozen() {
		return clienttypes.ErrClientFrozen
	}

	// value = sequence + path + acknowledgement
	value := append(
		combineSequenceAndPath(cs.ConsensusState.Sequence, path),
		acknowledgement...,
	)
	if cs.ConsensusState.PublicKey.VerifyBytes(value, proof) {
		return sdkerrors.Wrap(
			clienttypes.ErrFailedPacketAckVerification,
			"failed to verify proof against current public key, sequence, and acknowledgement",
		)
	}

	//cs.ConsensusState.Sequence += 1
	return nil

}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	height uint64,
	prefix commitmentexported.Prefix,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if cs.IsFrozen() {
		return clienttypes.ErrClientFrozen
	}

	// value = sequence + path
	value := combineSequenceAndPath(cs.ConsensusState.Sequence, path)

	if cs.ConsensusState.PublicKey.VerifyBytes(value, proof) {
		return sdkerrors.Wrap(
			clienttypes.ErrFailedPacketAckAbsenceVerification,
			"failed to verify proof against current public key, sequence, and an absent acknowledgement",
		)
	}

	//cs.ConsensusState.Sequence += 1
	return nil

}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	height uint64,
	prefix commitmentexported.Prefix,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	nextSequenceRecv uint64,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	if cs.IsFrozen() {
		return clienttypes.ErrClientFrozen
	}

	// value = sequence + path + nextSequenceRecv
	value := append(
		combineSequenceAndPath(cs.ConsensusState.Sequence, path),
		sdk.Uint64ToBigEndian(nextSequenceRecv)...,
	)

	if cs.ConsensusState.PublicKey.VerifyBytes(value, proof) {
		return sdkerrors.Wrap(
			clienttypes.ErrFailedNextSeqRecvVerification,
			"failed to verify proof against current public key, sequence, and the next sequence number to be received",
		)
	}

	//cs.ConsensusState.Sequence += 1
	return nil
}

// combineSequenceAndPath appends the sequence and path represented as bytes.
func combineSequenceAndPath(sequence uint64, path commitmenttypes.MerklePath) []byte {
	return append(
		sdk.Uint64ToBigEndian(sequence),
		[]byte(path.String())...,
	)
}
