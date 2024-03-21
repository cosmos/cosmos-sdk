package appmodule

import (
	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/transaction"
)

// Message aliases protoiface.MessageV1 for convenience.
type Message = transaction.Type

func messageName[M Message]() string {
	switch m := any(*new(M)).(type) {
	case gogoproto.Message:
		return gogoproto.MessageName(m)
	default:
		panic("unknown message type")
	}
}
