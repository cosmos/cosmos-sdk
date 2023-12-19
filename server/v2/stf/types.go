package stf

import "google.golang.org/protobuf/proto"

func TypeName(msg Type) string {
	return string(proto.MessageName(msg))
}
