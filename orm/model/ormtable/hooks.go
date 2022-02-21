package ormtable

import "google.golang.org/protobuf/proto"

// Hooks defines an interface for a table hooks which can intercept
// insert, update and delete operations. Table.Save and Table.Delete methods will
// do a type assertion on kvstore.IndexCommitmentStore and if the Hooks
// interface is defined call the appropriate hooks.
type Hooks interface {
	// OnInsert is called before the message is inserted.
	// If error is not nil the operation will fail.
	OnInsert(proto.Message) error

	// OnUpdate is called before the existing message is updated with the new one.
	// If error is not nil the operation will fail.
	OnUpdate(existing, new proto.Message) error

	// OnDelete is called before the message is deleted.
	// If error is not nil the operation will fail.
	OnDelete(proto.Message) error
}
