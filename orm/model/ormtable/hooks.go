package ormtable

import "google.golang.org/protobuf/proto"

type Hooks interface {
	OnInsert(proto.Message) error
	OnUpdate(existing, new proto.Message) error
	OnDelete(proto.Message) error
}
