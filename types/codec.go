package types

import (
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
)

const (
	// MsgInterfaceProtoName defines the protobuf name of the cosmos Msg interface
	MsgInterfaceProtoName = "cosmos.base.v1beta1.Msg"
)

// RegisterLegacyAminoCodec registers the sdk message type.
func RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	cdc.RegisterInterface((*transaction.Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}

// RegisterInterfaces registers the sdk message type.
func RegisterInterfaces(registry registry.InterfaceRegistrar) {
	registry.RegisterInterface(MsgInterfaceProtoName, (*Msg)(nil))
}
