package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// List of know interface types names
// TODO(fdymylja): maybe we need to add versioning to those constants?
const (
	MsgInterfaceName        = "cosmos.base.v1beta1.Msg"
	ServiceMsgInterfaceName = "cosmos.base.v1beta1.ServiceMsg"
)

// RegisterLegacyAminoCodec registers the sdk message type.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}

// RegisterInterfaces registers the sdk message type.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(MsgInterfaceName, (*Msg)(nil))
	// the interface name for MsgRequest is ServiceMsg because this is most useful for clients
	// to understand - it will be the way for clients to introspect on available Msg service methods
	registry.RegisterInterface(ServiceMsgInterfaceName, (*MsgRequest)(nil))
}
