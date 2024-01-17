package appmanager

import "google.golang.org/protobuf/proto"

type (
	Type     = proto.Message
	Identity = []byte
	Hash     = []byte
)

func TypeName(msg Type) string {
	return string(proto.MessageName(msg))
}
