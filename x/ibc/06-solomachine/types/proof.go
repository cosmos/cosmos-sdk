package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// CheckSignature verifies if the the provided public key generated the signature
// over the given data.
func CheckSignature(pubKey crypto.PubKey, data, signature []byte) error {
	if !pubKey.VerifyBytes(data, signature) {
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
		header.NewPubKey.Bytes()...,
	)
}

// ConsensusStateSignBytes returns the sign bytes for verification of the
// consensus state.
//
// Format: {sequence}{path}{consensus-state}
func ConsensusStateSignBytes(
	cdc *codec.Codec,
	sequence uint64,
	path commitmenttypes.MerklePath,
	consensusState ConsensusState,
) ([]byte, error) {
	bz, err := cdc.MarshalBinaryBare(consensusState)
	if err != nil {
		return nil, err
	}

	// sequence + path + consensus state
	return append(
		combineSequenceAndPath(sequence, path),
		bz...,
	), nil
}

// ConnectionStateSignBytes returns the sign bytes for verification of the
// connection state.
//
// Format: {sequence}{path}{connection-end}
func ConnectionStateSignBytes(
	cdc codec.Marshaler,
	sequence uint64,
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

	// sequence + path + connection end
	return append(
		combineSequenceAndPath(sequence, path),
		bz...,
	), nil
}

// ChannelStateSignBytes returns the sign bytes for verification of the
// channel state.
//
// Format: {sequence}{path}{channel-end}
func ChannelStateSignBytes(
	cdc codec.Marshaler,
	sequence uint64,
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

	// sequence + path + channel
	return append(
		combineSequenceAndPath(sequence, path),
		bz...,
	), nil
}

// PacketCommitmentSignBytes returns the sign bytes for verification of the
// packet commitment.
//
// Format: {sequence}{path}{commitment-bytes}
func PacketCommitmentSignBytes(
	sequence uint64,
	path commitmenttypes.MerklePath,
	commitmentBytes []byte,
) []byte {

	// sequence + path + commitment bytes
	return append(
		combineSequenceAndPath(sequence, path),
		commitmentBytes...,
	)
}

// PacketAcknowledgementSignBytes returns the sign bytes for verification of
// the acknowledgement.
//
// Format: {sequence}{path}{acknowledgement}
func PacketAcknowledgementSignBytes(
	sequence uint64,
	path commitmenttypes.MerklePath,
	acknowledgement []byte,
) []byte {

	// sequence + path + acknowledgement
	return append(
		combineSequenceAndPath(sequence, path),
		acknowledgement...,
	)
}

// PacketAcknowledgementAbsenceSignBytes returns the sign bytes for verificaiton
// of the absense of an acknowledgement.
//
// Format: {sequence}{path}
func PacketAcknowledgementAbsenceSignBytes(
	sequence uint64,
	path commitmenttypes.MerklePath,
) []byte {
	// value = sequence + path
	return combineSequenceAndPath(sequence, path)
}

// NextSequenceRecv returns the sign bytes for verification of the next
// sequence to be received.
//
// Format: {sequence}{path}{next-sequence-recv}
func NextSequenceRecvSignBytes(
	sequence uint64,
	path commitmenttypes.MerklePath,
	nextSequenceRecv uint64,
) []byte {

	// sequence + path + nextSequenceRecv
	return append(
		combineSequenceAndPath(sequence, path),
		sdk.Uint64ToBigEndian(nextSequenceRecv)...,
	)
}

// combineSequenceAndPath appends the sequence and path represented as bytes.
func combineSequenceAndPath(sequence uint64, path commitmenttypes.MerklePath) []byte {
	return append(
		sdk.Uint64ToBigEndian(sequence),
		[]byte(path.String())...,
	)
}
