package types

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strings"

	ics23 "github.com/confio/ics23/go"

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
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ clientexported.ClientState = (*ClientState)(nil)

// NewClientState creates a new ClientState instance
func NewClientState(chainID string, height clienttypes.Height) *ClientState {
	return &ClientState{
		ChainId: chainID,
		Height:  height,
	}
}

// GetChainID returns an empty string
func (cs ClientState) GetChainID() string {
	return cs.ChainId
}

// ClientType is localhost.
func (cs ClientState) ClientType() clientexported.ClientType {
	return clientexported.Localhost
}

// GetLatestHeight returns the latest height stored.
func (cs ClientState) GetLatestHeight() uint64 {
	return cs.Height.EpochHeight
}

// IsFrozen returns false.
func (cs ClientState) IsFrozen() bool {
	return false
}

// GetFrozenHeight returns 0.
func (cs ClientState) GetFrozenHeight() uint64 {
	return 0
}

// Validate performs a basic validation of the client state fields.
func (cs ClientState) Validate() error {
	if strings.TrimSpace(cs.ChainId) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidChainID, "chain id cannot be blank")
	}
	if cs.Height.EpochHeight == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "local epoch height cannot be zero")
	}
	return nil
}

// GetProofSpecs returns nil since localhost does not have to verify proofs
func (cs ClientState) GetProofSpecs() []*ics23.ProofSpec {
	return nil
}

// CheckHeaderAndUpdateState updates the localhost client. It only needs access to the context
// Localhost client does not have the logic override implemented. Even though we don't use the flag
// override, we still need to add it in order for the clientStatus interface to be respected.
func (cs ClientState) CheckHeaderAndUpdateState(
	ctx sdk.Context, _ codec.BinaryMarshaler, _ sdk.KVStore, _ clientexported.Header, override bool,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	// Hardcode 0 for epoch number for now
	// TODO: Retrieve epoch number from chain-id
	return NewClientState(
		ctx.ChainID(), // use the chain ID from context since the client is from the running chain (i.e self).
		clienttypes.NewHeight(0, uint64(ctx.BlockHeight())),
	), nil, nil
}

// CheckMisbehaviourAndUpdateState implements ClientState
// Since localhost is the client of the running chain, misbehaviour cannot be submitted to it
// Thus, CheckMisbehaviourAndUpdateState returns an error for localhost
func (cs ClientState) CheckMisbehaviourAndUpdateState(
	_ sdk.Context, _ codec.BinaryMarshaler, _ sdk.KVStore, _ clientexported.Misbehaviour,
) (clientexported.ClientState, error) {
	return nil, sdkerrors.Wrap(clienttypes.ErrInvalidMisbehaviour, "cannot submit misbehaviour to localhost client")
}

// VerifyClientState verifies that the localhost client state is stored locally
func (cs ClientState) VerifyClientState(
	store sdk.KVStore, cdc codec.BinaryMarshaler, _ commitmentexported.Root,
	_ uint64, _ commitmentexported.Prefix, _ string, _ []byte, clientState clientexported.ClientState,
) error {
	path := host.KeyClientState()
	bz := store.Get(path)
	if bz == nil {
		return sdkerrors.Wrapf(clienttypes.ErrFailedClientStateVerification,
			"not found for path: %s", path)
	}

	selfClient := clienttypes.MustUnmarshalClientState(cdc, bz)

	if !reflect.DeepEqual(selfClient, clientState) {
		return sdkerrors.Wrapf(clienttypes.ErrFailedClientStateVerification,
			"stored clientState != provided clientState: \n%v\n≠\n%v",
			selfClient, clientState,
		)
	}
	return nil
}

// VerifyClientConsensusState returns nil since a local host client does not store consensus
// states.
func (cs ClientState) VerifyClientConsensusState(
	sdk.KVStore, codec.BinaryMarshaler, commitmentexported.Root,
	uint64, string, uint64, commitmentexported.Prefix, []byte, clientexported.ConsensusState,
) error {
	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored locally.
func (cs ClientState) VerifyConnectionState(
	store sdk.KVStore,
	cdc codec.BinaryMarshaler,
	_ uint64,
	_ commitmentexported.Prefix,
	_ []byte,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
) error {
	path := host.KeyConnection(connectionID)
	bz := store.Get(path)
	if bz == nil {
		return sdkerrors.Wrapf(clienttypes.ErrFailedConnectionStateVerification, "not found for path %s", path)
	}

	var prevConnection connectiontypes.ConnectionEnd
	err := cdc.UnmarshalBinaryBare(bz, &prevConnection)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(&prevConnection, connectionEnd) {
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
	cdc codec.BinaryMarshaler,
	_ uint64,
	prefix commitmentexported.Prefix,
	_ []byte,
	portID,
	channelID string,
	channel channelexported.ChannelI,
) error {
	path := host.KeyChannel(portID, channelID)
	bz := store.Get(path)
	if bz == nil {
		return sdkerrors.Wrapf(clienttypes.ErrFailedChannelStateVerification, "not found for path %s", path)
	}

	var prevChannel channeltypes.Channel
	err := cdc.UnmarshalBinaryBare(bz, &prevChannel)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(&prevChannel, channel) {
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
	_ codec.BinaryMarshaler,
	_ uint64,
	_ commitmentexported.Prefix,
	_ []byte,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
) error {
	path := host.KeyPacketCommitment(portID, channelID, sequence)

	data := store.Get(path)
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
	_ codec.BinaryMarshaler,
	_ uint64,
	_ commitmentexported.Prefix,
	_ []byte,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
) error {
	path := host.KeyPacketAcknowledgement(portID, channelID, sequence)

	data := store.Get(path)
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
	_ codec.BinaryMarshaler,
	_ uint64,
	_ commitmentexported.Prefix,
	_ []byte,
	portID,
	channelID string,
	sequence uint64,
) error {
	path := host.KeyPacketAcknowledgement(portID, channelID, sequence)

	data := store.Get(path)
	if data != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckAbsenceVerification, "expected no ack absence")
	}

	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	store sdk.KVStore,
	_ codec.BinaryMarshaler,
	_ uint64,
	_ commitmentexported.Prefix,
	_ []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) error {
	path := host.KeyNextSequenceRecv(portID, channelID)

	data := store.Get(path)
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
