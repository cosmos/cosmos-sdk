package appmodule

import (
	gogoproto "github.com/cosmos/gogoproto/proto"
)

type Message = gogoproto.Message

func messageName[M Message]() string {
	switch m := any(*new(M)).(type) {
	case gogoproto.Message:
		return gogoproto.MessageName(m)
	default:
		panic("unknown message type")
	}
}
