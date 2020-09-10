package types

import (
	ics23 "github.com/confio/ics23/go"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

var _ exported.ClientState = (*ClientState)(nil)

// NewClientState creates a new ClientState instance.
func NewClientState(latestSequence uint64, consensusState *ConsensusState, allowUpdateAfterProposal bool) *ClientState {
	return &ClientState{
		Sequence:                 latestSequence,
		FrozenSequence:           0,
		ConsensusState:           consensusState,
		AllowUpdateAfterProposal: allowUpdateAfterProposal,
	}
}

// ClientType is Solo Machine.
func (cs ClientState) ClientType() exported.ClientType {
	return exported.SoloMachine
}

// GetLatestHeight returns the latest sequence number.
// Return exported.Height to satisfy ClientState interface
// Epoch number is always 0 for a solo-machine.
func (cs ClientState) GetLatestHeight() exported.Height {
	return clienttypes.NewHeight(0, cs.Sequence)
}

// IsFrozen returns true if the client is frozen.
func (cs ClientState) IsFrozen() bool {
	return cs.FrozenSequence != 0
}

// GetFrozenHeight returns the frozen sequence of the client.
// Return exported.Height to satisfy interface
// Epoch number is always 0 for a solo-machine
func (cs ClientState) GetFrozenHeight() exported.Height {
	return clienttypes.NewHeight(0, cs.FrozenSequence)
}

// GetProofSpecs returns nil proof specs since client state verification uses signatures.
func (cs ClientState) GetProofSpecs() []*ics23.ProofSpec {
	return nil
}

// Validate performs basic validation of the client state fields.
func (cs ClientState) Validate() error {
	if cs.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidClient, "sequence cannot be 0")
	}
	if cs.ConsensusState == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be nil")
	}
	return cs.ConsensusState.ValidateBasic()
}

// VerifyClientState verifies a proof of the client state of the running chain
// stored on the solo machine.
func (cs ClientState) VerifyClientState(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	_ exported.Root,
	height exported.Height,
	prefix exported.Prefix,
	counterpartyClientIdentifier string,
	proof []byte,
	clientState exported.ClientState,
) error {
	signature, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ClientStatePath()
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	signBz, err := ClientStateSignBytes(cdc, sequence, signature.Timestamp, cs.ConsensusState.Diversifier, path, clientState)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, signature.Signature); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyClientConsensusState verifies a proof of the consensus state of the
// running chain stored on the solo machine.
func (cs ClientState) VerifyClientConsensusState(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	_ exported.Root,
	height exported.Height,
	counterpartyClientIdentifier string,
	consensusHeight exported.Height,
	prefix exported.Prefix,
	proof []byte,
	consensusState exported.ConsensusState,
) error {
	signature, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	signBz, err := ConsensusStateSignBytes(cdc, sequence, signature.Timestamp, cs.ConsensusState.Diversifier, path, consensusState)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, signature.Signature); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (cs ClientState) VerifyConnectionState(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	connectionID string,
	connectionEnd exported.ConnectionI,
) error {
	signature, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	signBz, err := ConnectionStateSignBytes(cdc, sequence, signature.Timestamp, cs.ConsensusState.Diversifier, path, connectionEnd)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, signature.Signature); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (cs ClientState) VerifyChannelState(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	channel exported.ChannelI,
) error {
	signature, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	signBz, err := ChannelStateSignBytes(cdc, sequence, signature.Timestamp, cs.ConsensusState.Diversifier, path, channel)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, signature.Signature); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	commitmentBytes []byte,
) error {
	signature, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	signBz, err := PacketCommitmentSignBytes(cdc, sequence, signature.Timestamp, cs.ConsensusState.Diversifier, path, commitmentBytes)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, signature.Signature); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	acknowledgement []byte,
) error {
	signature, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	signBz, err := PacketAcknowledgementSignBytes(cdc, sequence, signature.Timestamp, cs.ConsensusState.Diversifier, path, acknowledgement)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, signature.Signature); err != nil {
		return err
	}

	cs.Sequence++
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
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
) error {
	signature, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	signBz, err := PacketAcknowledgementAbsenceSignBytes(cdc, sequence, signature.Timestamp, cs.ConsensusState.Diversifier, path)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, signature.Signature); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) error {
	signature, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	signBz, err := NextSequenceRecvSignBytes(cdc, sequence, signature.Timestamp, cs.ConsensusState.Diversifier, path, nextSequenceRecv)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, signature.Signature); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// produceVerificationArgs perfoms the basic checks on the arguments that are
// shared between the verification functions and returns the unmarshalled
// proof representing the signature and timestamp along with the solo-machine sequence
// encoded in the proofHeight.
func produceVerificationArgs(
	cdc codec.BinaryMarshaler,
	cs ClientState,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
) (signature TimestampedSignature, sequence uint64, err error) {
	if epoch := height.GetEpochNumber(); epoch != 0 {
		return TimestampedSignature{}, 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "epoch must be 0 for solomachine, got epoch-number: %d", epoch)
	}
	// sequence is encoded in the epoch height of height struct
	sequence = height.GetEpochHeight()
	if cs.IsFrozen() {
		return TimestampedSignature{}, 0, clienttypes.ErrClientFrozen
	}

	if prefix == nil {
		return TimestampedSignature{}, 0, sdkerrors.Wrap(commitmenttypes.ErrInvalidPrefix, "prefix cannot be empty")
	}

	_, ok := prefix.(commitmenttypes.MerklePrefix)
	if !ok {
		return TimestampedSignature{}, 0, sdkerrors.Wrapf(commitmenttypes.ErrInvalidPrefix, "invalid prefix type %T, expected MerklePrefix", prefix)
	}

	if proof == nil {
		return TimestampedSignature{}, 0, sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof cannot be empty")
	}

	if err = cdc.UnmarshalBinaryBare(proof, &signature); err != nil {
		return TimestampedSignature{}, 0, sdkerrors.Wrapf(ErrInvalidProof, "failed to unmarshal proof into type %T", TimestampedSignature{})
	}

	if cs.ConsensusState == nil {
		return TimestampedSignature{}, 0, sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be empty")
	}

	latestSequence := cs.GetLatestHeight().GetEpochHeight()
	if latestSequence < sequence {
		return TimestampedSignature{}, 0, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"client state sequence < proof sequence (%d < %d)", latestSequence, sequence,
		)
	}

	if cs.ConsensusState.GetTimestamp() > signature.Timestamp {
		return TimestampedSignature{}, 0, sdkerrors.Wrapf(ErrInvalidProof, "the consensus state timestamp is greater than the signature timestamp (%d >= %d)", cs.ConsensusState.GetTimestamp(), signature.Timestamp)
	}

	return signature, sequence, nil
}

// sets the client state to the store
func setClientState(store sdk.KVStore, cdc codec.BinaryMarshaler, clientState exported.ClientState) {
	bz := clienttypes.MustMarshalClientState(cdc, clientState)
	store.Set(host.KeyClientState(), bz)
}
