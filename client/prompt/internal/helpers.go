package internal

import (
	cosmos_proto "github.com/cosmos/cosmos-proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
)

const (
	AddressStringScalarType          = "cosmos.AddressString"
	ValidatorAddressStringScalarType = "cosmos.ValidatorAddressString"
	ConsensusAddressStringScalarType = "cosmos.ConsensusAddressString"
)

// GetScalarType gets scalar type of a field.
// Copied from client/v2/flag package. Lives here to avoid circular dependency.
func GetScalarType(field protoreflect.FieldDescriptor) (string, bool) {
	scalar := proto.GetExtension(field.Options(), cosmos_proto.E_Scalar)
	scalarStr, ok := scalar.(string)
	return scalarStr, ok
}

// GetSignerFieldName gets signer field name of a message.
// AutoCLI supports only one signer field per message.
// Copied from client/v2/flag package. Lives here to avoid circular dependency.
func GetSignerFieldName(descriptor protoreflect.MessageDescriptor) string {
	signersFields := proto.GetExtension(descriptor.Options(), msgv1.E_Signer).([]string)
	if len(signersFields) == 0 {
		return ""
	}

	return signersFields[0]
}
