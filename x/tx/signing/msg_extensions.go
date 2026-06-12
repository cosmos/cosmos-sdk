package signing

import (
	"google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

// E_Service and E_Signer are protov2-compatible (protoimpl.ExtensionInfo)
// definitions of the cosmos.msg.v1 proto extensions. Defining them here avoids
// importing cosmossdk.io/api/cosmos/msg/v1 from x/tx/signing, which would
// create the cycle x/tx/signing→types/msgservice→x/tx/signing.
//
// Field numbers and semantics are identical to the pulsar-generated versions.
// The file descriptor is registered by the gogoproto types/msgservice init().
var (
	E_Service = &protoimpl.ExtensionInfo{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         11110000,
		Name:          "cosmos.msg.v1.service",
		Tag:           "varint,11110000,opt,name=service",
		Filename:      "cosmos/msg/v1/msg.proto",
	}

	E_Signer = &protoimpl.ExtensionInfo{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*[]string)(nil),
		Field:         11110000,
		Name:          "cosmos.msg.v1.signer",
		Tag:           "bytes,11110000,rep,name=signer",
		Filename:      "cosmos/msg/v1/msg.proto",
	}
)
