package appmodule

import (
	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
)

// MsgTypeURL returns the TypeURL of a proto message.
// Note, this function adds `/` to the message name.
func MsgTypeURL(msg proto.Message) string {
	if m, ok := msg.(protov2.Message); ok {
		return "/" + string(m.ProtoReflect().Descriptor().FullName())
	}

	return "/" + proto.MessageName(msg)
}
