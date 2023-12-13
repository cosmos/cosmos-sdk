package appmanager

import "google.golang.org/protobuf/proto"

type Type = proto.Message
type Identity = []byte
type Hash = []byte

func TypeName(msg Type) string {
	return string(proto.MessageName(msg))
}

type Event = any
