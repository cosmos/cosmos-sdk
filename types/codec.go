package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

const (
	// MsgInterfaceProtoName defines the protobuf name of the cosmos Msg interface
	MsgInterfaceProtoName = "cosmos.base.v1beta1.Msg"
	// ServiceMsgInterfaceProtoName defines the protobuf name of the cosmos MsgRequest interface
	ServiceMsgInterfaceProtoName = "cosmos.base.v1beta1.ServiceMsg"
)

// RegisterLegacyAminoCodec registers the sdk message type.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}

// RegisterInterfaces registers the sdk message type.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(MsgInterfaceProtoName, (*Msg)(nil))
	// the interface name for MsgRequest is ServiceMsg because this is most useful for clients
	// to understand - it will be the way for clients to introspect on available Msg service methods
	registry.RegisterInterface(ServiceMsgInterfaceProtoName, (*MsgRequest)(nil))
}
