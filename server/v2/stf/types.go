package stf

import (
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/core/transaction"
)

func typeName(msg transaction.Type) string {
	return string(proto.MessageName(msg))
}
