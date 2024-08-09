package app

import (
	gogoproto "cosmossdk.io/core/transaction"
)

// MsgInterfaceProtoName defines the protobuf name of the cosmos Msg interface
const MsgInterfaceProtoName = "cosmos.base.v1beta1.Msg"

type ProtoCodec interface {
	Marshal(v gogoproto.Msg) ([]byte, error)
	Unmarshal(data []byte, v gogoproto.Msg) error
	Name() string
}

type InterfaceRegistry interface {
	AnyResolver
	ListImplementations(ifaceTypeURL string) []string
	ListAllInterfaces() []string
}

type AnyResolver = interface {
	Resolve(typeUrl string) (gogoproto.Msg, error)
}
