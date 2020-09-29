package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// RegisterInterfaces register the ibc channel submodule interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*exported.ClientState)(nil),
		&ClientState{},
	)
	registry.RegisterImplementations(
		(*exported.ConsensusState)(nil),
		&ConsensusState{},
	)
	registry.RegisterImplementations(
		(*exported.Header)(nil),
		&Header{},
	)
	registry.RegisterImplementations(
		(*exported.Misbehaviour)(nil),
		&Misbehaviour{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/light-clients/solomachine module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc/light-clients/solomachine and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

func UnmarshalSignatureData(cdc codec.BinaryMarshaler, data []byte) (signing.SignatureData, error) {
	protoSigData := &signing.SignatureDescriptor_Data{}
	if err := cdc.UnmarshalBinaryBare(data, protoSigData); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to unmarshal proof into type %T", protoSigData)
	}

	sigData := signing.SignatureDataFromProto(protoSigData)

	return sigData, nil
}

// CanUnmarshalDataByType returns true if the data provided can be unmarshaled
// to the specified DataType.
func CanUnmarshalDataByType(cdc codec.BinaryMarshaler, dataType DataType, data []byte) bool {
	if len(data) == 0 {
		return false
	}

	switch dataType {
	case UNSPECIFIED:
		return false

	case CLIENT:
		clientData := &ClientStateData{}
		if err := cdc.UnmarshalBinaryBare(data, clientData); err != nil {
			return false
		}

		// unpack any
		if _, err := clienttypes.UnpackClientState(clientData.ClientState); err != nil {
			fmt.Println(clientData.ClientState)
			fmt.Println(err)
			return false
		}
		return true

	case CONSENSUS:
		consensusData := &ConsensusStateData{}
		if err := cdc.UnmarshalBinaryBare(data, consensusData); err != nil {
			return false
		}

		// unpack any
		if _, err := clienttypes.UnpackConsensusState(consensusData.ConsensusState); err != nil {
			return false
		}
		return true

	case CONNECTION:
		connectionData := &ConnectionStateData{}
		return cdc.UnmarshalBinaryBare(data, connectionData) == nil

	case CHANNEL:
		channelData := &ChannelStateData{}
		return cdc.UnmarshalBinaryBare(data, channelData) == nil

	case PACKETCOMMITMENT:
		commitmentData := &PacketCommitmentData{}
		return cdc.UnmarshalBinaryBare(data, commitmentData) == nil

	case PACKETACKNOWLEDGEMENT:
		ackData := &PacketAcknowledgementData{}
		return cdc.UnmarshalBinaryBare(data, ackData) == nil

	case PACKETACKNOWLEDGEMENTABSENCE:
		ackAbsenceData := &PacketAcknowledgementAbsenceData{}
		return cdc.UnmarshalBinaryBare(data, ackAbsenceData) == nil

	case NEXTSEQUENCERECV:
		nextSeqRecvData := &NextSequenceRecvData{}
		return cdc.UnmarshalBinaryBare(data, nextSeqRecvData) == nil

	default:
		return false
	}
}
