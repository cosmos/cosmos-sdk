package stf

import (
	"cosmossdk.io/server/v2/core/transaction"
	"google.golang.org/protobuf/proto"
)

func typeName(msg transaction.Type) string {
	return string(proto.MessageName(msg))
}
