package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
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
