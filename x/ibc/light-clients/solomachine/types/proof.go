package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// VerifySignature verifies if the the provided public key generated the signature
// over the given data.
func VerifySignature(pubKey crypto.PubKey, data, signature []byte) error {
	if !pubKey.VerifySignature(data, signature) {
		return ErrSignatureVerificationFailed
	}

	return nil
}

// EvidenceSignBytes returns the sign bytes for verification of misbehaviour.
//
// Format: {sequence}{data}
func EvidenceSignBytes(sequence uint64, data []byte) []byte {
	return append(
		sdk.Uint64ToBigEndian(sequence),
		data...,
	)
}

// HeaderSignBytes returns the sign bytes for verification of misbehaviour.
//
// Format: {sequence}{header.newPubKey}
func HeaderSignBytes(header Header) []byte {
	return append(
		sdk.Uint64ToBigEndian(header.Sequence),
		header.GetPubKey().Bytes()...,
	)
}

// ClientStateSignBytes returns the sign bytes for verification of the
// client state.
//
// Format: {sequence}{timestamp}{path}{client-state}
func ClientStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	path commitmenttypes.MerklePath,
	clientState clientexported.ClientState,
) ([]byte, error) {
	bz, err := codec.MarshalAny(cdc, clientState)
	if err != nil {
		return nil, err
	}

	// sequence + timestamp + path + client state
	return append(
		combineSequenceTimestampPath(sequence, timestamp, path),
		bz...,
	), nil
}

// ConsensusStateSignBytes returns the sign bytes for verification of the
// consensus state.
//
// Format: {sequence}{timestamp}{path}{consensus-state}
func ConsensusStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	path commitmenttypes.MerklePath,
	consensusState *ConsensusState,
) ([]byte, error) {
	bz, err := codec.MarshalAny(cdc, consensusState)
	if err != nil {
		return nil, err
	}

	// sequence + timestamp + path + consensus state
	return append(
		combineSequenceTimestampPath(sequence, timestamp, path),
		bz...,
	), nil
}

// ConnectionStateSignBytes returns the sign bytes for verification of the
// connection state.
//
// Format: {sequence}{timestamp}{path}{connection-end}
func ConnectionStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	path commitmenttypes.MerklePath,
	connectionEnd connectionexported.ConnectionI,
) ([]byte, error) {
	connection, ok := connectionEnd.(connectiontypes.ConnectionEnd)
	if !ok {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "invalid connection type %T", connectionEnd)
	}

	bz, err := cdc.MarshalBinaryBare(&connection)
	if err != nil {
		return nil, err
	}

	// sequence + timestamp + path + connection end
	return append(
		combineSequenceTimestampPath(sequence, timestamp, path),
		bz...,
	), nil
}

// ChannelStateSignBytes returns the sign bytes for verification of the
// channel state.
//
// Format: {sequence}{timestamp}{path}{channel-end}
func ChannelStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	path commitmenttypes.MerklePath,
	channelEnd channelexported.ChannelI,
) ([]byte, error) {
	channel, ok := channelEnd.(channeltypes.Channel)
	if !ok {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "invalid channel type %T", channelEnd)
	}

	bz, err := cdc.MarshalBinaryBare(&channel)
	if err != nil {
		return nil, err
	}

	// sequence + timestamp + path + channel
	return append(
		combineSequenceTimestampPath(sequence, timestamp, path),
		bz...,
	), nil
}

// PacketCommitmentSignBytes returns the sign bytes for verification of the
// packet commitment.
//
// Format: {sequence}{timestamp}{path}{commitment-bytes}
func PacketCommitmentSignBytes(
	sequence, timestamp uint64,
	path commitmenttypes.MerklePath,
	commitmentBytes []byte,
) []byte {

	// sequence + timestamp + path + commitment bytes
	return append(
		combineSequenceTimestampPath(sequence, timestamp, path),
		commitmentBytes...,
	)
}

// PacketAcknowledgementSignBytes returns the sign bytes for verification of
// the acknowledgement.
//
// Format: {sequence}{timestamp}{path}{acknowledgement}
func PacketAcknowledgementSignBytes(
	sequence, timestamp uint64,
	path commitmenttypes.MerklePath,
	acknowledgement []byte,
) []byte {

	// sequence + timestamp + path + acknowledgement
	return append(
		combineSequenceTimestampPath(sequence, timestamp, path),
		acknowledgement...,
	)
}

// PacketAcknowledgementAbsenceSignBytes returns the sign bytes for verification
// of the absence of an acknowledgement.
//
// Format: {sequence}{timestamp}{path}
func PacketAcknowledgementAbsenceSignBytes(
	sequence, timestamp uint64,
	path commitmenttypes.MerklePath,
) []byte {
	// value = sequence + timestamp + path
	return combineSequenceTimestampPath(sequence, timestamp, path)
}

// NextSequenceRecv returns the sign bytes for verification of the next
// sequence to be received.
//
// Format: {sequence}{timestamp}{path}{next-sequence-recv}
func NextSequenceRecvSignBytes(
	sequence, timestamp uint64,
	path commitmenttypes.MerklePath,
	nextSequenceRecv uint64,
) []byte {

	// sequence + timestamp + path + nextSequenceRecv
	return append(
		combineSequenceTimestampPath(sequence, timestamp, path),
		sdk.Uint64ToBigEndian(nextSequenceRecv)...,
	)
}

// combineSequenceTimestampPath combines the sequence, the timestamp and
// the path into one byte slice.
func combineSequenceTimestampPath(sequence, timestamp uint64, path commitmenttypes.MerklePath) []byte {
	bz := append(sdk.Uint64ToBigEndian(sequence), sdk.Uint64ToBigEndian(timestamp)...)
	return append(
		bz,
		[]byte(path.String())...,
	)
}
