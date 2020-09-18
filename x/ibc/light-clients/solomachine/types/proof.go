package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// VerifySignature verifies if the the provided public key generated the signature
// over the given data.
func VerifySignature(pubKey crypto.PubKey, data, signature []byte) error {
	if !pubKey.VerifySignature(data, signature) {
		return ErrSignatureVerificationFailed
	}

	return nil
}

// MisbehaviourSignBytes returns the sign bytes for verification of misbehaviour.
func MisbehaviourSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	data []byte) ([]byte, error) {
	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        data,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// HeaderSignBytes returns the sign bytes for verification of misbehaviour.
func HeaderSignBytes(
	cdc codec.BinaryMarshaler,
	header *Header,
) ([]byte, error) {
	data := &HeaderData{
		NewPubKey:      header.NewPublicKey,
		NewDiversifier: header.NewDiversifier,
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    header.Sequence,
		Timestamp:   header.Timestamp,
		Diversifier: header.NewDiversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// ClientStateSignBytes returns the sign bytes for verification of the
// client state.
func ClientStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath, // nolint: interfacer
	clientState exported.ClientState,
) ([]byte, error) {
	any, err := clienttypes.PackClientState(clientState)
	if err != nil {
		return nil, err
	}

	data := &ClientStateData{
		Path:        []byte(path.String()),
		ClientState: any,
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// ConsensusStateSignBytes returns the sign bytes for verification of the
// consensus state.
func ConsensusStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath, // nolint: interfacer
	consensusState exported.ConsensusState,
) ([]byte, error) {
	any, err := clienttypes.PackConsensusState(consensusState)
	if err != nil {
		return nil, err
	}

	data := &ConsensusStateData{
		Path:           []byte(path.String()),
		ConsensusState: any,
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// ConnectionStateSignBytes returns the sign bytes for verification of the
// connection state.
func ConnectionStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath, // nolint: interfacer
	connectionEnd exported.ConnectionI,
) ([]byte, error) {
	connection, ok := connectionEnd.(connectiontypes.ConnectionEnd)
	if !ok {
		return nil, sdkerrors.Wrapf(
			connectiontypes.ErrInvalidConnection,
			"expected type %T, got %T", connectiontypes.ConnectionEnd{}, connectionEnd,
		)
	}

	data := &ConnectionStateData{
		Path:       []byte(path.String()),
		Connection: &connection,
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// ChannelStateSignBytes returns the sign bytes for verification of the
// channel state.
func ChannelStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath, // nolint: interfacer
	channelEnd exported.ChannelI,
) ([]byte, error) {
	channel, ok := channelEnd.(channeltypes.Channel)
	if !ok {
		return nil, sdkerrors.Wrapf(
			channeltypes.ErrInvalidChannel,
			"expected channel type %T, got %T", channeltypes.Channel{}, channelEnd)
	}

	data := &ChannelStateData{
		Path:    []byte(path.String()),
		Channel: &channel,
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// PacketCommitmentSignBytes returns the sign bytes for verification of the
// packet commitment.
func PacketCommitmentSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath, // nolint: interfacer
	commitmentBytes []byte,
) ([]byte, error) {
	data := &PacketCommitmentData{
		Path:       []byte(path.String()),
		Commitment: commitmentBytes,
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// PacketAcknowledgementSignBytes returns the sign bytes for verification of
// the acknowledgement.
func PacketAcknowledgementSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath, // nolint: interfacer
	acknowledgement []byte,
) ([]byte, error) {
	data := &PacketAcknowledgementData{
		Path:            []byte(path.String()),
		Acknowledgement: acknowledgement,
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// PacketAcknowledgementAbsenceSignBytes returns the sign bytes for verification
// of the absence of an acknowledgement.
func PacketAcknowledgementAbsenceSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath, // nolint: interfacer
) ([]byte, error) {
	data := &PacketAcknowledgementAbsenseData{
		Path: []byte(path.String()),
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// NextSequenceRecvSignBytes returns the sign bytes for verification of the next
// sequence to be received.
func NextSequenceRecvSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath, // nolint: interfacer
	nextSequenceRecv uint64,
) ([]byte, error) {
	data := &NextSequenceRecvData{
		Path:        []byte(path.String()),
		NextSeqRecv: nextSequenceRecv,
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}
