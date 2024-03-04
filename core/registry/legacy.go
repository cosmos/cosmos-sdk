package registry

import gogoproto "github.com/cosmos/gogoproto/proto"

type LegacyRegistry interface {
	RegisterInterface(protoName string, iface interface{}, impls ...gogoproto.Message)

	// RegisterImplementations registers impls as concrete implementations of
	// the interface iface.
	//
	// Ex:
	//  registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSend{}, &MsgMultiSend{})
	RegisterImplementations(iface interface{}, impls ...gogoproto.Message)
}
