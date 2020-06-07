package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

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

// ClientState requires (read-only) access to keys outside the client prefix.
type ClientState struct {
	ID      string `json:"id" yaml:"id"`
	ChainID string `json:"chain_id" yaml:"chain_id"`
	Height  int64  `json:"height" yaml:"height"`
}

// NewClientState creates a new ClientState instance
func NewClientState(chainID string, height int64) ClientState {
	return ClientState{
		ID:      clientexported.Localhost.String(),
		ChainID: chainID,
		Height:  height,
	}
}

// GetID returns the loop-back client state identifier.
func (cs ClientState) GetID() string {
	return cs.ID
}

// GetChainID returns an empty string
func (cs ClientState) GetChainID() string {
	return cs.ChainID
}

// ClientType is localhost.
func (cs ClientState) ClientType() clientexported.ClientType {
	return clientexported.Localhost
}

// GetLatestHeight returns the latest height stored.
func (cs ClientState) GetLatestHeight() uint64 {
	return uint64(cs.Height)
}

// IsFrozen returns false.
func (cs ClientState) IsFrozen() bool {
	return false
}

// Validate performs a basic validation of the client state fields.
func (cs ClientState) Validate() error {
	if strings.TrimSpace(cs.ChainID) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidChainID, "chain id cannot be blank")
	}
	if cs.Height <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "height must be positive: %d", cs.Height)
	}
	return host.ClientIdentifierValidator(cs.ID)
}

// VerifyClientConsensusState verifies a proof of the consensus
// state of the loop-back client.
// VerifyClientConsensusState verifies a proof of the consensus state of the
// Tendermint client stored on the target machine.
func (cs ClientState) VerifyClientConsensusState(
	store sdk.KVStore,
	cdc *codec.Codec,
	_ commitmentexported.Root,
	height uint64,
	_ string,
	consensusHeight uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, consensusStatePath(cs.GetID()))
	if err != nil {
		return err
	}

	data := store.Get([]byte(path.String()))
	if len(data) == 0 {
		return sdkerrors.Wrapf(clienttypes.ErrFailedClientConsensusStateVerification, "not found for path %s", path)
	}

	var prevConsensusState clientexported.ConsensusState
	if err := cdc.UnmarshalBinaryBare(data, &prevConsensusState); err != nil {
		return err
	}

	if consensusState != prevConsensusState {
		return sdkerrors.Wrapf(
			clienttypes.ErrFailedClientConsensusStateVerification,
			"consensus state ≠ previous stored consensus state: \n%v\n≠\n%v", consensusState, prevConsensusState,
		)
	}

	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored locally.
func (cs ClientState) VerifyConnectionState(
	store sdk.KVStore,
	cdc codec.Marshaler,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	bz := store.Get([]byte(path.String()))
	if bz == nil {
		return sdkerrors.Wrapf(clienttypes.ErrFailedConnectionStateVerification, "not found for path %s", path)
	}

	var prevConnection connectiontypes.ConnectionEnd
	err = cdc.UnmarshalBinaryBare(bz, &prevConnection)
	if err != nil {
		return err
	}

	if connectionEnd != &prevConnection {
		return sdkerrors.Wrapf(
			clienttypes.ErrFailedConnectionStateVerification,
			"connection end ≠ previous stored connection: \n%v\n≠\n%v", connectionEnd, prevConnection,
		)
	}

	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the local machine.
func (cs ClientState) VerifyChannelState(
	store sdk.KVStore,
	cdc codec.Marshaler,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	channel channelexported.ChannelI,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	bz := store.Get([]byte(path.String()))
	if bz == nil {
		return sdkerrors.Wrapf(clienttypes.ErrFailedChannelStateVerification, "not found for path %s", path)
	}

	var prevChannel channeltypes.Channel
	err = cdc.UnmarshalBinaryBare(bz, &prevChannel)
	if err != nil {
		return err
	}

	if channel != &prevChannel {
		return sdkerrors.Wrapf(
			clienttypes.ErrFailedChannelStateVerification,
			"channel end ≠ previous stored channel: \n%v\n≠\n%v", channel, prevChannel,
		)
	}

	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	store sdk.KVStore,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	data := store.Get([]byte(path.String()))
	if len(data) == 0 {
		return sdkerrors.Wrapf(clienttypes.ErrFailedPacketCommitmentVerification, "not found for path %s", path)
	}

	if !bytes.Equal(data, commitmentBytes) {
		return sdkerrors.Wrapf(
			clienttypes.ErrFailedPacketCommitmentVerification,
			"commitment ≠ previous commitment: \n%X\n≠\n%X", commitmentBytes, data,
		)
	}

	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	store sdk.KVStore,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	data := store.Get([]byte(path.String()))
	if len(data) == 0 {
		return sdkerrors.Wrapf(clienttypes.ErrFailedPacketAckVerification, "not found for path %s", path)
	}

	if !bytes.Equal(data, acknowledgement) {
		return sdkerrors.Wrapf(
			clienttypes.ErrFailedPacketAckVerification,
			"ak bytes ≠ previous ack: \n%X\n≠\n%X", acknowledgement, data,
		)
	}

	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	store sdk.KVStore,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	data := store.Get([]byte(path.String()))
	if data != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckAbsenceVerification, "expected no ack absence")
	}

	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	store sdk.KVStore,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	nextSequenceRecv uint64,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	data := store.Get([]byte(path.String()))
	if len(data) == 0 {
		return sdkerrors.Wrapf(clienttypes.ErrFailedNextSeqRecvVerification, "not found for path %s", path)
	}

	prevSequenceRecv := binary.BigEndian.Uint64(data)
	if prevSequenceRecv != nextSequenceRecv {
		return sdkerrors.Wrapf(
			clienttypes.ErrFailedNextSeqRecvVerification,
			"next sequence receive ≠ previous stored sequence (%d ≠ %d)", nextSequenceRecv, prevSequenceRecv,
		)
	}

	return nil
}

// consensusStatePath takes an Identifier and returns a Path under which to
// store the consensus state of a client.
func consensusStatePath(clientID string) string {
	return fmt.Sprintf("consensusState/%s", clientID)
}
