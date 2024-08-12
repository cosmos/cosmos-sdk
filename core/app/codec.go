package app

import (
	"cosmossdk.io/core/transaction"
)

// MsgInterfaceProtoName defines the protobuf name of the cosmos Msg interface
const MsgInterfaceProtoName = "cosmos.base.v1beta1.Msg"

type ProtoCodec interface {
	Marshal(v transaction.Msg) ([]byte, error)
	Unmarshal(data []byte, v transaction.Msg) error
	Name() string
}

type InterfaceRegistry interface {
	AnyResolver
	ListImplementations(ifaceTypeURL string) []string
	ListAllInterfaces() []string
}

type AnyResolver = interface {
	Resolve(typeUrl string) (transaction.Msg, error)
}
