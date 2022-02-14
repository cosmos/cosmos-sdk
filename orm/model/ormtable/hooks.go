package ormtable

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// ValidateHooks defines an interface for a table hooks which can intercept
// insert, update and delete operations. Table.Save and Table.Delete methods will
// do a type assertion on kvstore.IndexCommitmentStore and if the ValidateHooks
// interface is defined call the appropriate hooks.
type ValidateHooks interface {
	// ValidateInsert is called before the message is inserted.
	// If error is not nil the operation will fail.
	ValidateInsert(context.Context, proto.Message) error

	// ValidateUpdate is called before the existing message is updated with the new one.
	// If error is not nil the operation will fail.
	ValidateUpdate(ctx context.Context, existing, new proto.Message) error

	// ValidateDelete is called before the message is deleted.
	// If error is not nil the operation will fail.
	ValidateDelete(context.Context, proto.Message) error
}

type WriteHooks interface {
	OnInsert(context.Context, proto.Message)
	OnUpdate(ctx context.Context, existing, new proto.Message)
	OnDelete(context.Context, proto.Message)
}
