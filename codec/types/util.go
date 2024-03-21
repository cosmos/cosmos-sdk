package types

import (
	"github.com/cosmos/gogoproto/proto"
)

// MsgTypeURL returns the TypeURL of a `sdk.Msg`.
func MsgTypeURL(msg proto.Message) string {
	return "/" + proto.MessageName(msg)
}
