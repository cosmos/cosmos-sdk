package appmodule

import (
	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
)

type Message = gogoproto.Message

func MessageName[M Message]() string {
	switch m := any(*new(M)).(type) {
	case protov2.Message:
		return string(m.ProtoReflect().Descriptor().FullName())
	case gogoproto.Message:
		return gogoproto.MessageName(m)
	default:
		panic("unknown message type")
	}
}
