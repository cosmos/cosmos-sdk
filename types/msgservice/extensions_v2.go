package msgservice

import (
	"google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

// E_ServiceV2 and E_SignerV2 are protov2-compatible (protoimpl.ExtensionInfo)
// equivalents of E_Service and E_Signer. Use these with
// google.golang.org/protobuf/proto.HasExtension / proto.GetExtension.
//
// The underlying proto field numbers and semantics are identical to the
// pulsar-generated cosmossdk.io/api/cosmos/msg/v1 extensions; the file
// descriptor is registered by the gogoproto msg.pb.go init().
var (
	E_ServiceV2 = &protoimpl.ExtensionInfo{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         11110000,
		Name:          "cosmos.msg.v1.service",
		Tag:           "varint,11110000,opt,name=service",
		Filename:      "cosmos/msg/v1/msg.proto",
	}

	E_SignerV2 = &protoimpl.ExtensionInfo{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*[]string)(nil),
		Field:         11110000,
		Name:          "cosmos.msg.v1.signer",
		Tag:           "bytes,11110000,rep,name=signer",
		Filename:      "cosmos/msg/v1/msg.proto",
	}
)
