package app

import (
	"github.com/cosmos/gogoproto/jsonpb"
	gogoproto "github.com/cosmos/gogoproto/proto"
)

// MsgInterfaceProtoName defines the protobuf name of the cosmos Msg interface
const MsgInterfaceProtoName = "cosmos.base.v1beta1.Msg"

type ProtoCodec interface {
	Marshal(v gogoproto.Message) ([]byte, error)
	Unmarshal(data []byte, v gogoproto.Message) error
	Name() string
}

type InterfaceRegistry interface {
	jsonpb.AnyResolver
	ListImplementations(ifaceTypeURL string) []string
	ListAllInterfaces() []string
}
