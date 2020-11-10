package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// VerifySignature verifies if the the provided public key generated the signature
// over the given data. Single and Multi signature public keys are supported.
// The signature data type must correspond to the public key type. An error is
// returned if signature verification fails or an invalid SignatureData type is
// provided.
func VerifySignature(pubKey cryptotypes.PubKey, signBytes []byte, sigData signing.SignatureData) error {
	switch pubKey := pubKey.(type) {
	case multisig.PubKey:
		data, ok := sigData.(*signing.MultiSignatureData)
		if !ok {
			return sdkerrors.Wrapf(ErrSignatureVerificationFailed, "invalid signature data type, expected %T, got %T", (*signing.MultiSignatureData)(nil), data)
		}

		// The function supplied fulfills the VerifyMultisignature interface. No special
		// adjustments need to be made to the sign bytes based on the sign mode.
		if err := pubKey.VerifyMultisignature(func(signing.SignMode) ([]byte, error) {
			return signBytes, nil
		}, data); err != nil {
			return err
		}

	default:
		data, ok := sigData.(*signing.SingleSignatureData)
		if !ok {
			return sdkerrors.Wrapf(ErrSignatureVerificationFailed, "invalid signature data type, expected %T, got %T", (*signing.SingleSignatureData)(nil), data)
		}

		if !pubKey.VerifySignature(signBytes, data.Signature) {
			return ErrSignatureVerificationFailed
		}
	}

	return nil
}

// MisbehaviourSignBytes returns the sign bytes for verification of misbehaviour.
func MisbehaviourSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	dataType DataType,
	data []byte) ([]byte, error) {
	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    dataType,
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
		DataType:    HEADER,
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
	path commitmenttypes.MerklePath,
	clientState exported.ClientState,
) ([]byte, error) {
	dataBz, err := ClientStateDataBytes(cdc, path, clientState)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    CLIENT,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// ClientStateDataBytes returns the client state data bytes used in constructing
// SignBytes.
func ClientStateDataBytes(
	cdc codec.BinaryMarshaler,
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

	return dataBz, nil
}

// ConsensusStateSignBytes returns the sign bytes for verification of the
// consensus state.
func ConsensusStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath,
	consensusState exported.ConsensusState,
) ([]byte, error) {
	dataBz, err := ConsensusStateDataBytes(cdc, path, consensusState)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    CONSENSUS,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// ConsensusStateDataBytes returns the consensus state data bytes used in constructing
// SignBytes.
func ConsensusStateDataBytes(
	cdc codec.BinaryMarshaler,
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

	return dataBz, nil
}

// ConnectionStateSignBytes returns the sign bytes for verification of the
// connection state.
func ConnectionStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath,
	connectionEnd exported.ConnectionI,
) ([]byte, error) {
	dataBz, err := ConnectionStateDataBytes(cdc, path, connectionEnd)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    CONNECTION,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// ConnectionStateDataBytes returns the connection state data bytes used in constructing
// SignBytes.
func ConnectionStateDataBytes(
	cdc codec.BinaryMarshaler,
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

	return dataBz, nil
}

// ChannelStateSignBytes returns the sign bytes for verification of the
// channel state.
func ChannelStateSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath,
	channelEnd exported.ChannelI,
) ([]byte, error) {
	dataBz, err := ChannelStateDataBytes(cdc, path, channelEnd)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    CHANNEL,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// ChannelStateDataBytes returns the channel state data bytes used in constructing
// SignBytes.
func ChannelStateDataBytes(
	cdc codec.BinaryMarshaler,
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

	return dataBz, nil
}

// PacketCommitmentSignBytes returns the sign bytes for verification of the
// packet commitment.
func PacketCommitmentSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath,
	commitmentBytes []byte,
) ([]byte, error) {
	dataBz, err := PacketCommitmentDataBytes(cdc, path, commitmentBytes)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    PACKETCOMMITMENT,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// PacketCommitmentDataBytes returns the packet commitment data bytes used in constructing
// SignBytes.
func PacketCommitmentDataBytes(
	cdc codec.BinaryMarshaler,
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

	return dataBz, nil
}

// PacketAcknowledgementSignBytes returns the sign bytes for verification of
// the acknowledgement.
func PacketAcknowledgementSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath,
	acknowledgement []byte,
) ([]byte, error) {
	dataBz, err := PacketAcknowledgementDataBytes(cdc, path, acknowledgement)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    PACKETACKNOWLEDGEMENT,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// PacketAcknowledgementDataBytes returns the packet acknowledgement data bytes used in constructing
// SignBytes.
func PacketAcknowledgementDataBytes(
	cdc codec.BinaryMarshaler,
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

	return dataBz, nil
}

// PacketReceiptAbsenceSignBytes returns the sign bytes for verification
// of the absence of an receipt.
func PacketReceiptAbsenceSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath,
) ([]byte, error) {
	dataBz, err := PacketReceiptAbsenceDataBytes(cdc, path)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    PACKETRECEIPTABSENCE,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// PacketReceiptAbsenceDataBytes returns the packet receipt absence data bytes
// used in constructing SignBytes.
func PacketReceiptAbsenceDataBytes(
	cdc codec.BinaryMarshaler,
	path commitmenttypes.MerklePath, // nolint: interfacer
) ([]byte, error) {
	data := &PacketReceiptAbsenceData{
		Path: []byte(path.String()),
	}

	dataBz, err := cdc.MarshalBinaryBare(data)
	if err != nil {
		return nil, err
	}

	return dataBz, nil
}

// NextSequenceRecvSignBytes returns the sign bytes for verification of the next
// sequence to be received.
func NextSequenceRecvSignBytes(
	cdc codec.BinaryMarshaler,
	sequence, timestamp uint64,
	diversifier string,
	path commitmenttypes.MerklePath,
	nextSequenceRecv uint64,
) ([]byte, error) {
	dataBz, err := NextSequenceRecvDataBytes(cdc, path, nextSequenceRecv)
	if err != nil {
		return nil, err
	}

	signBytes := &SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		DataType:    NEXTSEQUENCERECV,
		Data:        dataBz,
	}

	return cdc.MarshalBinaryBare(signBytes)
}

// NextSequenceRecvDataBytes returns the next sequence recv data bytes used in constructing
// SignBytes.
func NextSequenceRecvDataBytes(
	cdc codec.BinaryMarshaler,
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

	return dataBz, nil
}
