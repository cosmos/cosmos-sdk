package ormtable

import "google.golang.org/protobuf/proto"

// Hooks defines an interface for a table hooks which can intercept
// insert, update and delete operations.
type Hooks interface {
	OnInsert(proto.Message) error
	OnUpdate(existing, new proto.Message) error
	OnDelete(proto.Message) error
}
