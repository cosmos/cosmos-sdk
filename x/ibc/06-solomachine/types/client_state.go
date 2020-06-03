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
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
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
	return NewClientState(msg.GetClientID(), msg.ConsensusState), nil
}

// NewClientState creates a new ClientState instance.
func NewClientState(id string, consensusState ConsensusState) ClientState {
	return ClientState{
		ID:             id,
		Frozen:         false,
		ConsensusState: consensusState,
	}
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

// Validate performs basic validation of the client state fields.
func (cs ClientState) Validate() error {
	if err := host.ClientIdentifierValidator(cs.ID); err != nil {
		return err
	}
	return cs.ConsensusState.ValidateBasic()
}

// VerifyClientConsensusState verifies a proof of the consensus state of the
// Solo Machine client stored on the target machine.
func (cs ClientState) VerifyClientConsensusState(
	store sdk.KVStore,
	cdc *codec.Codec,
	root commitmentexported.Root,
	sequence uint64,
	counterpartyClientIdentifier string,
	consensusHeight uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	_ clientexported.ConsensusState,
) error {
	if err := validateVerificationArgs(cs, sequence, prefix, proof, cs.ConsensusState); err != nil {
		return err
	}

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	data, err := ConsensusStateSignBytes(cdc, sequence, path, cs.ConsensusState)
	if err != nil {
		return err
	}

	if err := CheckSignature(cs.ConsensusState.PubKey, data, proof); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedClientConsensusStateVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	setClientState(store, cs)
	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (cs ClientState) VerifyConnectionState(
	store sdk.KVStore,
	cdc codec.Marshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
	_ clientexported.ConsensusState,
) error {
	if err := validateVerificationArgs(cs, sequence, prefix, proof, cs.ConsensusState); err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	data, err := ConnectionStateSignBytes(cdc, sequence, path, connectionEnd)
	if err != nil {
		return err
	}

	if err := CheckSignature(cs.ConsensusState.PubKey, data, proof); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedConnectionStateVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	setClientState(store, cs)
	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (cs ClientState) VerifyChannelState(
	store sdk.KVStore,
	cdc codec.Marshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	channel channelexported.ChannelI,
	_ clientexported.ConsensusState,
) error {
	if err := validateVerificationArgs(cs, sequence, prefix, proof, cs.ConsensusState); err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	data, err := ChannelStateSignBytes(cdc, sequence, path, channel)
	if err != nil {
		return err
	}

	if err := CheckSignature(cs.ConsensusState.PubKey, data, proof); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedChannelStateVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	setClientState(store, cs)
	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	store sdk.KVStore,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	commitmentBytes []byte,
	_ clientexported.ConsensusState,
) error {
	if err := validateVerificationArgs(cs, sequence, prefix, proof, cs.ConsensusState); err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketCommitmentSignBytes(sequence, path, commitmentBytes)

	if err := CheckSignature(cs.ConsensusState.PubKey, data, proof); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketCommitmentVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	setClientState(store, cs)
	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	store sdk.KVStore,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	acknowledgement []byte,
	_ clientexported.ConsensusState,
) error {
	if err := validateVerificationArgs(cs, sequence, prefix, proof, cs.ConsensusState); err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketAcknowledgementSignBytes(sequence, path, acknowledgement)

	if err := CheckSignature(cs.ConsensusState.PubKey, data, proof); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	setClientState(store, cs)
	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	store sdk.KVStore,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	_ clientexported.ConsensusState,
) error {
	if err := validateVerificationArgs(cs, sequence, prefix, proof, cs.ConsensusState); err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketAcknowledgementAbsenceSignBytes(sequence, path)

	if err := CheckSignature(cs.ConsensusState.PubKey, data, proof); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckAbsenceVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	setClientState(store, cs)
	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	store sdk.KVStore,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
	_ clientexported.ConsensusState,
) error {
	if err := validateVerificationArgs(cs, sequence, prefix, proof, cs.ConsensusState); err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	data := NextSequenceRecvSignBytes(sequence, path, nextSequenceRecv)

	if err := CheckSignature(cs.ConsensusState.PubKey, data, proof); err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrFailedNextSeqRecvVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	setClientState(store, cs)
	return nil
}

// validateVerificationArgs perfoms the basic checks on the arguments that are
// shared between the verification functions.
func validateVerificationArgs(
	cs ClientState,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	consensusState clientexported.ConsensusState,
) error {
	if cs.GetLatestHeight() < sequence {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"client state (%s) sequence < proof sequence (%d < %d)", cs.ID, cs.GetLatestHeight(), sequence,
		)
	}

	if cs.IsFrozen() {
		return clienttypes.ErrClientFrozen
	}

	if prefix == nil {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidPrefix, "prefix cannot be empty")
	}

	_, ok := prefix.(commitmenttypes.MerklePrefix)
	if !ok {
		return sdkerrors.Wrapf(commitmenttypes.ErrInvalidPrefix, "invalid prefix type %T, expected MerklePrefix", prefix)
	}

	if proof == nil {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof cannot be empty")
	}

	if consensusState == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be empty")
	}

	_, ok = consensusState.(ConsensusState)
	if !ok {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "invalid consensus type %T, expected %T", consensusState, ConsensusState{})
	}

	return nil
}

// sets the client state to the store
func setClientState(store sdk.KVStore, clientState clientexported.ClientState) {
	bz := SubModuleCdc.MustMarshalBinaryBare(clientState)
	store.Set(host.KeyClientState(), bz)
}
