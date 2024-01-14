package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// MsgTypeURL returns the TypeURL of a `sdk.Msg`.
func MsgTypeURL[T sdk.ProtoMessage](v T) string {
	switch msg := any(v).(type) {
	case sdk.Msg:
		return "/" + proto.MessageName(msg)
	case sdk.MsgV2:
		return "/" + string(msg.ProtoReflect().Descriptor().FullName())
	default:
		return ""
	}
}
