package types

import (
	ics23 "github.com/confio/ics23/go"

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

var _ clientexported.ClientState = (*ClientState)(nil)

// NewClientState creates a new ClientState instance.
func NewClientState(id, chainID string, consensusState *ConsensusState) *ClientState {
	return &ClientState{
		Id:             id,
		ChainId:        chainID,
		FrozenHeight:   0,
		ConsensusState: consensusState,
	}
}

// GetID returns the solo machine client state identifier.
func (cs ClientState) GetID() string {
	return cs.Id
}

// GetChainID returns an empty string.
func (cs ClientState) GetChainID() string {
	return cs.ChainId
}

// ClientType is Solo Machine.
func (cs ClientState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetLatestHeight returns the latest sequence number.
func (cs ClientState) GetLatestHeight() uint64 {
	return cs.ConsensusState.Sequence
}

// IsFrozen returns true if the client is frozen.
func (cs ClientState) IsFrozen() bool {
	return cs.FrozenHeight != 0
}

// GetFrozenHeight returns the frozen height of the client.
func (cs ClientState) GetFrozenHeight() uint64 {
	return cs.FrozenHeight
}

// GetProofSpecs returns nil proof specs since client state verification uses signatures.
func (cs ClientState) GetProofSpecs() []*ics23.ProofSpec {
	return nil
}

// Validate performs basic validation of the client state fields.
func (cs ClientState) Validate() error {
	if err := host.ClientIdentifierValidator(cs.Id); err != nil {
		return err
	}
	return cs.ConsensusState.ValidateBasic()
}

// VerifyClientState TODO
func (cs ClientState) VerifyClientState(
	store sdk.KVStore, cdc codec.BinaryMarshaler, _ commitmentexported.Root,
	_ uint64, _ commitmentexported.Prefix, _ string, _ []byte, clientState clientexported.ClientState,
) error {
	return nil
}

// VerifyClientConsensusState verifies a proof of the consensus state of the
// Solo Machine client stored on the target machine.
func (cs ClientState) VerifyClientConsensusState(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	root commitmentexported.Root,
	sequence uint64,
	counterpartyClientIdentifier string,
	consensusHeight uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	_ clientexported.ConsensusState,
) error {
	signature, err := produceVerificationArgs(cdc, cs, sequence, prefix, proof)
	if err != nil {
		return err
	}

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	data, err := ConsensusStateSignBytes(cdc, sequence, signature.Timestamp, path, cs.ConsensusState)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return err
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (cs ClientState) VerifyConnectionState(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
) error {
	signature, err := produceVerificationArgs(cdc, cs, sequence, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	data, err := ConnectionStateSignBytes(cdc, sequence, signature.Timestamp, path, connectionEnd)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return err
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (cs ClientState) VerifyChannelState(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	channel channelexported.ChannelI,
) error {
	signature, err := produceVerificationArgs(cdc, cs, sequence, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	data, err := ChannelStateSignBytes(cdc, sequence, signature.Timestamp, path, channel)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return err
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	commitmentBytes []byte,
) error {
	signature, err := produceVerificationArgs(cdc, cs, sequence, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketCommitmentSignBytes(sequence, signature.Timestamp, path, commitmentBytes)

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return err
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	acknowledgement []byte,
) error {
	signature, err := produceVerificationArgs(cdc, cs, sequence, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketAcknowledgementSignBytes(sequence, signature.Timestamp, path, acknowledgement)

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return err
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
) error {
	signature, err := produceVerificationArgs(cdc, cs, sequence, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketAcknowledgementAbsenceSignBytes(sequence, signature.Timestamp, path)

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return err
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) error {
	signature, err := produceVerificationArgs(cdc, cs, sequence, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	data := NextSequenceRecvSignBytes(sequence, signature.Timestamp, path, nextSequenceRecv)

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return err
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// produceVerificationArgs perfoms the basic checks on the arguments that are
// shared between the verification functions and returns the unmarshalled
// proof representing the signature and timestamp.
func produceVerificationArgs(
	cdc codec.BinaryMarshaler,
	cs ClientState,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
) (signature Signature, err error) {
	if cs.IsFrozen() {
		return Signature{}, clienttypes.ErrClientFrozen
	}

	if prefix == nil {
		return Signature{}, sdkerrors.Wrap(commitmenttypes.ErrInvalidPrefix, "prefix cannot be empty")
	}

	_, ok := prefix.(commitmenttypes.MerklePrefix)
	if !ok {
		return Signature{}, sdkerrors.Wrapf(commitmenttypes.ErrInvalidPrefix, "invalid prefix type %T, expected MerklePrefix", prefix)
	}

	if proof == nil {
		return Signature{}, sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof cannot be empty")
	}

	if err = cdc.UnmarshalBinaryBare(proof, &signature); err != nil {
		return Signature{}, sdkerrors.Wrapf(ErrInvalidProof, "failed to unmarshal proof into type %T", Signature{})
	}

	if cs.ConsensusState == nil {
		return Signature{}, sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be empty")
	}

	if cs.GetLatestHeight() < sequence {
		return Signature{}, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"client state (%s) sequence < proof sequence (%d < %d)", cs.Id, cs.GetLatestHeight(), sequence,
		)
	}

	if cs.ConsensusState.GetTimestamp() > signature.Timestamp {
		return Signature{}, sdkerrors.Wrapf(ErrInvalidProof, "the consensus state timestamp is greater than the signature timestamp (%d >= %d)", cs.ConsensusState.GetTimestamp(), signature.Timestamp)
	}

	return signature, nil
}

// sets the client state to the store
func setClientState(store sdk.KVStore, cdc codec.BinaryMarshaler, clientState clientexported.ClientState) {
	bz, err := codec.MarshalAny(cdc, clientState)
	if err != nil {
		panic(err)
	}
	store.Set(host.KeyClientState(), bz)
}
