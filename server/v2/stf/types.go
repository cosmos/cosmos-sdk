package stf

import (
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/server/v2/core/transaction"
)

func typeName(msg transaction.Type) string {
	return string(proto.MessageName(msg))
}
