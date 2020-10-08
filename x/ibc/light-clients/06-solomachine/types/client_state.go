package types

import (
	ics23 "github.com/confio/ics23/go"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var _ exported.ClientState = (*ClientState)(nil)

// SoloMachine is used to indicate that the light client is a solo machine.
const SoloMachine string = "Solo Machine"

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
func (cs ClientState) ClientType() string {
	return SoloMachine
}

// GetLatestHeight returns the latest sequence number.
// Return exported.Height to satisfy ClientState interface
// Version number is always 0 for a solo-machine.
func (cs ClientState) GetLatestHeight() exported.Height {
	return clienttypes.NewHeight(0, cs.Sequence)
}

// IsFrozen returns true if the client is frozen.
func (cs ClientState) IsFrozen() bool {
	return cs.FrozenSequence != 0
}

// GetFrozenHeight returns the frozen sequence of the client.
// Return exported.Height to satisfy interface
// Version number is always 0 for a solo-machine
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

// ZeroCustomFields returns solomachine client state with client-specific fields FrozenSequence,
// and AllowUpdateAfterProposal zeroed out
func (cs ClientState) ZeroCustomFields() exported.ClientState {
	return NewClientState(
		cs.Sequence, cs.ConsensusState, false,
	)
}

// VerifyUpgrade returns an error since solomachine client does not support upgrades
func (cs ClientState) VerifyUpgrade(
	_ sdk.Context, _ codec.BinaryMarshaler, _ sdk.KVStore,
	_ exported.ClientState, _ exported.Height, _ []byte,
) error {
	return sdkerrors.Wrap(clienttypes.ErrInvalidUpgradeClient, "cannot upgrade solomachine client")
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
	sigData, timestamp, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ClientStatePath()
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	signBz, err := ClientStateSignBytes(cdc, sequence, timestamp, cs.ConsensusState.Diversifier, path, clientState)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, sigData); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = timestamp
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
	sigData, timestamp, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	signBz, err := ConsensusStateSignBytes(cdc, sequence, timestamp, cs.ConsensusState.Diversifier, path, consensusState)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, sigData); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = timestamp
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
	sigData, timestamp, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	signBz, err := ConnectionStateSignBytes(cdc, sequence, timestamp, cs.ConsensusState.Diversifier, path, connectionEnd)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, sigData); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = timestamp
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
	sigData, timestamp, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	signBz, err := ChannelStateSignBytes(cdc, sequence, timestamp, cs.ConsensusState.Diversifier, path, channel)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, sigData); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = timestamp
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
	sigData, timestamp, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	signBz, err := PacketCommitmentSignBytes(cdc, sequence, timestamp, cs.ConsensusState.Diversifier, path, commitmentBytes)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, sigData); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = timestamp
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
	sigData, timestamp, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	signBz, err := PacketAcknowledgementSignBytes(cdc, sequence, timestamp, cs.ConsensusState.Diversifier, path, acknowledgement)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, sigData); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = timestamp
	setClientState(store, cdc, &cs)
	return nil
}

// VerifyPacketReceiptAbsence verifies a proof of the absence of an
// incoming packet receipt at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketReceiptAbsence(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
) error {
	sigData, timestamp, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketReceiptPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	signBz, err := PacketReceiptAbsenceSignBytes(cdc, sequence, timestamp, cs.ConsensusState.Diversifier, path)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, sigData); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = timestamp
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
	sigData, timestamp, sequence, err := produceVerificationArgs(cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	signBz, err := NextSequenceRecvSignBytes(cdc, sequence, timestamp, cs.ConsensusState.Diversifier, path, nextSequenceRecv)
	if err != nil {
		return err
	}

	if err := VerifySignature(cs.ConsensusState.GetPubKey(), signBz, sigData); err != nil {
		return err
	}

	cs.Sequence++
	cs.ConsensusState.Timestamp = timestamp
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
) (signing.SignatureData, uint64, uint64, error) {
	if version := height.GetVersionNumber(); version != 0 {
		return nil, 0, 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "version must be 0 for solomachine, got version-number: %d", version)
	}
	// sequence is encoded in the version height of height struct
	sequence := height.GetVersionHeight()
	if cs.IsFrozen() {
		return nil, 0, 0, clienttypes.ErrClientFrozen
	}

	if prefix == nil {
		return nil, 0, 0, sdkerrors.Wrap(commitmenttypes.ErrInvalidPrefix, "prefix cannot be empty")
	}

	_, ok := prefix.(commitmenttypes.MerklePrefix)
	if !ok {
		return nil, 0, 0, sdkerrors.Wrapf(commitmenttypes.ErrInvalidPrefix, "invalid prefix type %T, expected MerklePrefix", prefix)
	}

	if proof == nil {
		return nil, 0, 0, sdkerrors.Wrap(ErrInvalidProof, "proof cannot be empty")
	}

	timestampedSigData := &TimestampedSignatureData{}
	if err := cdc.UnmarshalBinaryBare(proof, timestampedSigData); err != nil {
		return nil, 0, 0, sdkerrors.Wrapf(err, "failed to unmarshal proof into type %T", timestampedSigData)
	}

	timestamp := timestampedSigData.Timestamp

	if len(timestampedSigData.SignatureData) == 0 {
		return nil, 0, 0, sdkerrors.Wrap(ErrInvalidProof, "signature data cannot be empty")
	}

	sigData, err := UnmarshalSignatureData(cdc, timestampedSigData.SignatureData)
	if err != nil {
		return nil, 0, 0, err
	}

	if cs.ConsensusState == nil {
		return nil, 0, 0, sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be empty")
	}

	latestSequence := cs.GetLatestHeight().GetVersionHeight()
	if latestSequence < sequence {
		return nil, 0, 0, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"client state sequence < proof sequence (%d < %d)", latestSequence, sequence,
		)
	}

	if cs.ConsensusState.GetTimestamp() > timestamp {
		return nil, 0, 0, sdkerrors.Wrapf(ErrInvalidProof, "the consensus state timestamp is greater than the signature timestamp (%d >= %d)", cs.ConsensusState.GetTimestamp(), timestamp)
	}

	return sigData, timestamp, sequence, nil
}

// sets the client state to the store
func setClientState(store sdk.KVStore, cdc codec.BinaryMarshaler, clientState exported.ClientState) {
	bz := clienttypes.MustMarshalClientState(cdc, clientState)
	store.Set(host.KeyClientState(), bz)
}
