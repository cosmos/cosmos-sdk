package types

import (
	"strings"
	"time"

	ics23 "github.com/confio/ics23/go"
	lite "github.com/tendermint/tendermint/lite2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ clientexported.ClientState = ClientState{}

// InitializeFromMsg creates a tendermint client state from a CreateClientMsg
func InitializeFromMsg(msg *MsgCreateClient) ClientState {
	return NewClientState(msg.Header.ChainID, msg.TrustLevel,
		msg.TrustingPeriod, msg.UnbondingPeriod, msg.MaxClockDrift,
		uint64(msg.Header.Height), msg.ProofSpecs,
	)
}

// NewClientState creates a new ClientState instance
func NewClientState(
	chainID string, trustLevel Fraction,
	trustingPeriod, ubdPeriod, maxClockDrift time.Duration,
	latestHeight uint64, specs []*ics23.ProofSpec,
) ClientState {
	return ClientState{
		ChainID:         chainID,
		TrustLevel:      trustLevel,
		TrustingPeriod:  trustingPeriod,
		UnbondingPeriod: ubdPeriod,
		MaxClockDrift:   maxClockDrift,
		LatestHeight:    latestHeight,
		FrozenHeight:    0,
		ProofSpecs:      specs,
	}
}

// GetChainID returns the chain-id
func (cs ClientState) GetChainID() string {
	return cs.ChainID
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

// FrozenHeight returns the height at which client is frozen
// NOTE: FrozenHeight is 0 if client is unfrozen
func (cs ClientState) GetFrozenHeight() uint64 {
	return cs.FrozenHeight
}

// Validate performs a basic validation of the client state fields.
func (cs ClientState) Validate() error {
	if strings.TrimSpace(cs.ChainID) == "" {
		return sdkerrors.Wrap(ErrInvalidChainID, "chain id cannot be empty string")
	}
	if err := lite.ValidateTrustLevel(cs.TrustLevel.ToTendermint()); err != nil {
		return err
	}
	if cs.TrustingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidTrustingPeriod, "trusting period cannot be zero")
	}
	if cs.UnbondingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidUnbondingPeriod, "unbonding period cannot be zero")
	}
	if cs.MaxClockDrift == 0 {
		return sdkerrors.Wrap(ErrInvalidMaxClockDrift, "max clock drift cannot be zero")
	}
	if cs.LatestHeight == 0 {
		return sdkerrors.Wrap(ErrInvalidHeaderHeight, "tendermint height cannot be zero")
	}
	if cs.TrustingPeriod >= cs.UnbondingPeriod {
		return sdkerrors.Wrapf(
			ErrInvalidTrustingPeriod,
			"trusting period (%s) should be < unbonding period (%s)", cs.TrustingPeriod, cs.UnbondingPeriod,
		)
	}
	// Validate ProofSpecs
	if cs.ProofSpecs == nil {
		return sdkerrors.Wrap(ErrInvalidProofSpecs, "proof specs cannot be nil for tm client")
	}
	for _, spec := range cs.ProofSpecs {
		if spec == nil {
			return sdkerrors.Wrap(ErrInvalidProofSpecs, "proof spec cannot be nil")
		}
	}

	return nil
}

// GetProofSpecs returns the format the client expects for proof verification
// as a string array specifying the proof type for each position in chained proof
func (cs ClientState) GetProofSpecs() []*ics23.ProofSpec {
	return cs.ProofSpecs
}

// VerifyClientConsensusState verifies a proof of the consensus state of the
// Tendermint client stored on the target machine.
func (cs ClientState) VerifyClientConsensusState(
	_ sdk.KVStore,
	cdc codec.BinaryMarshaler,
	aminoCdc *codec.Codec,
	provingRoot commitmentexported.Root,
	height uint64,
	counterpartyClientIdentifier string,
	consensusHeight uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	consensusState clientexported.ConsensusState,
) error {
	merkleProof, err := sanitizeVerificationArgs(cdc, cs, height, prefix, proof, consensusState)
	if err != nil {
		return err
	}

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	bz, err := aminoCdc.MarshalBinaryBare(consensusState)
	if err != nil {
		return err
	}

	if err := merkleProof.VerifyMembership(cs.ProofSpecs, provingRoot, path, bz); err != nil {
		return err
	}

	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (cs ClientState) VerifyConnectionState(
	_ sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
	consensusState clientexported.ConsensusState,
) error {
	merkleProof, err := sanitizeVerificationArgs(cdc, cs, height, prefix, proof, consensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	connection, ok := connectionEnd.(connectiontypes.ConnectionEnd)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "invalid connection type %T", connectionEnd)
	}

	bz, err := cdc.MarshalBinaryBare(&connection)
	if err != nil {
		return err
	}

	if err := merkleProof.VerifyMembership(cs.ProofSpecs, consensusState.GetRoot(), path, bz); err != nil {
		return err
	}

	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (cs ClientState) VerifyChannelState(
	_ sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	channel channelexported.ChannelI,
	consensusState clientexported.ConsensusState,
) error {
	merkleProof, err := sanitizeVerificationArgs(cdc, cs, height, prefix, proof, consensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	channelEnd, ok := channel.(channeltypes.Channel)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "invalid channel type %T", channel)
	}

	bz, err := cdc.MarshalBinaryBare(&channelEnd)
	if err != nil {
		return err
	}

	if err := merkleProof.VerifyMembership(cs.ProofSpecs, consensusState.GetRoot(), path, bz); err != nil {
		return err
	}

	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	_ sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
	consensusState clientexported.ConsensusState,
) error {
	merkleProof, err := sanitizeVerificationArgs(cdc, cs, height, prefix, proof, consensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if err := merkleProof.VerifyMembership(cs.ProofSpecs, consensusState.GetRoot(), path, commitmentBytes); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketCommitmentVerification, err.Error())
	}

	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	_ sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
	consensusState clientexported.ConsensusState,
) error {
	merkleProof, err := sanitizeVerificationArgs(cdc, cs, height, prefix, proof, consensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if err := merkleProof.VerifyMembership(cs.ProofSpecs, consensusState.GetRoot(), path, channeltypes.CommitAcknowledgement(acknowledgement)); err != nil {
		return err
	}

	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	_ sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
	consensusState clientexported.ConsensusState,
) error {
	merkleProof, err := sanitizeVerificationArgs(cdc, cs, height, prefix, proof, consensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if err := merkleProof.VerifyNonMembership(cs.ProofSpecs, consensusState.GetRoot(), path); err != nil {
		return err
	}

	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	_ sdk.KVStore,
	cdc codec.BinaryMarshaler,
	height uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
	consensusState clientexported.ConsensusState,
) error {
	merkleProof, err := sanitizeVerificationArgs(cdc, cs, height, prefix, proof, consensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	bz := sdk.Uint64ToBigEndian(nextSequenceRecv)

	if err := merkleProof.VerifyMembership(cs.ProofSpecs, consensusState.GetRoot(), path, bz); err != nil {
		return err
	}

	return nil
}

// sanitizeVerificationArgs perfoms the basic checks on the arguments that are
// shared between the verification functions and returns the unmarshalled
// merkle proof and an error if one occurred.
func sanitizeVerificationArgs(
	cdc codec.BinaryMarshaler,
	cs ClientState,
	height uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	consensusState clientexported.ConsensusState,
) (merkleProof commitmenttypes.MerkleProof, err error) {
	if cs.GetLatestHeight() < height {
		return commitmenttypes.MerkleProof{}, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"client state height < proof height (%d < %d)", cs.GetLatestHeight(), height,
		)
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return commitmenttypes.MerkleProof{}, clienttypes.ErrClientFrozen
	}

	if prefix == nil {
		return commitmenttypes.MerkleProof{}, sdkerrors.Wrap(commitmenttypes.ErrInvalidPrefix, "prefix cannot be empty")
	}

	_, ok := prefix.(*commitmenttypes.MerklePrefix)
	if !ok {
		return commitmenttypes.MerkleProof{}, sdkerrors.Wrapf(commitmenttypes.ErrInvalidPrefix, "invalid prefix type %T, expected *MerklePrefix", prefix)
	}

	if proof == nil {
		return commitmenttypes.MerkleProof{}, sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof cannot be empty")
	}

	if err = cdc.UnmarshalBinaryBare(proof, &merkleProof); err != nil {
		return commitmenttypes.MerkleProof{}, sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "failed to unmarshal proof into commitment merkle proof")
	}

	if consensusState == nil {
		return commitmenttypes.MerkleProof{}, sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be empty")
	}

	_, ok = consensusState.(ConsensusState)
	if !ok {
		return commitmenttypes.MerkleProof{}, sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "invalid consensus type %T, expected %T", consensusState, ConsensusState{})
	}

	return merkleProof, nil
}
