package localhost

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ clientexported.ClientState = ClientState{}

// ClientState requires (read-only) access to keys outside the client prefix.
type ClientState struct {
	store   types.KVStore
	ID      string `json:"id" yaml:"id"`
	ChainID string `json:"chain_id" yaml:"chain_id"`
	Height  int64  `json:"height" yaml:"height"`
}

// NewClientState creates a new ClientState instance
func NewClientState(store types.KVStore, chainID string, height int64) ClientState {
	return ClientState{
		store:   store,
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

// VerifyClientConsensusState verifies a proof of the consensus
// state of the loop-back client.
// VerifyClientConsensusState verifies a proof of the consensus state of the
// Tendermint client stored on the target machine.
func (cs ClientState) VerifyClientConsensusState(
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

	data := cs.store.Get([]byte(path.String()))
	if len(data) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrFailedClientConsensusStateVerification, "not found")
	}

	var prevConsensusState exported.ConsensusState
	cdc.MustUnmarshalBinaryBare(data, &prevConsensusState)
	if consensusState != prevConsensusState {
		return sdkerrors.Wrap(clienttypes.ErrFailedClientConsensusStateVerification, "not equal")
	}

	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored locally.
func (cs ClientState) VerifyConnectionState(
	cdc *codec.Codec,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	bz := cs.store.Get([]byte(path.String()))
	if bz == nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedConnectionStateVerification, "not found")
	}

	var prevConnectionState connectionexported.ConnectionI
	cdc.MustUnmarshalBinaryBare(bz, &prevConnectionState)
	if connectionEnd != prevConnectionState {
		return sdkerrors.Wrap(clienttypes.ErrFailedConnectionStateVerification, "not equal")
	}

	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the local machine.
func (cs ClientState) VerifyChannelState(
	cdc *codec.Codec,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	channel channelexported.ChannelI,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	bz := cs.store.Get([]byte(path.String()))
	if bz == nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedChannelStateVerification, "not found")
	}

	var prevChannelState channelexported.ChannelI
	cdc.MustUnmarshalBinaryBare(bz, &prevChannelState)
	if channel != prevChannelState {
		return sdkerrors.Wrap(clienttypes.ErrFailedChannelStateVerification, "not equal")
	}

	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.PacketCommitmentPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	data := cs.store.Get([]byte(path.String()))
	if len(data) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketCommitmentVerification, "not found")
	}

	if !bytes.Equal(data, commitmentBytes) {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketCommitmentVerification, "not equal")
	}

	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	data := cs.store.Get([]byte(path.String()))
	if len(data) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckVerification, "not found")
	}

	if !bytes.Equal(data, acknowledgement) {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckVerification, "not equal")
	}

	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	data := cs.store.Get([]byte(path.String()))
	if data != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckAbsenceVerification, "expected no ack absence")
	}

	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	_ uint64,
	prefix commitmentexported.Prefix,
	_ commitmentexported.Proof,
	portID,
	channelID string,
	nextSequenceRecv uint64,
	_ clientexported.ConsensusState,
) error {
	path, err := commitmenttypes.ApplyPrefix(prefix, ibctypes.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	data := cs.store.Get([]byte(path.String()))
	if len(data) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrFailedNextSeqRecvVerification, "not found")
	}

	prevSequenceRecv := binary.BigEndian.Uint64(data)
	if prevSequenceRecv != nextSequenceRecv {
		return sdkerrors.Wrap(clienttypes.ErrFailedNextSeqRecvVerification, "not equal")
	}

	return nil
}

// consensusStatePath takes an Identifier and returns a Path under which to
// store the consensus state of a client.
func consensusStatePath(clientID string) string {
	return fmt.Sprintf("consensusState/%s", clientID)
}
