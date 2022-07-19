package ormtable

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// ValidateHooks defines an interface for a table hooks which can intercept
// insert, update and delete operations and possibly return an error.
type ValidateHooks interface {
	// ValidateInsert is called before the message is inserted.
	// If error is not nil the insertion will fail.
	ValidateInsert(context.Context, proto.Message) error

	// ValidateUpdate is called before the existing message is updated with the new one.
	// If error is not nil the update will fail.
	ValidateUpdate(ctx context.Context, existing, new proto.Message) error

	// ValidateDelete is called before the message is deleted.
	// If error is not nil the deletion will fail.
	ValidateDelete(context.Context, proto.Message) error
}

// WriteHooks defines an interface for listening to insertions, updates and
// deletes after they are written to the store. This can be used for indexing
// state in another database. Indexers should make sure they coordinate with
// transactions at live at the next level above the ORM as they write hooks
// may be called but the enclosing transaction may still fail. The context
// is provided in each method to help coordinate this.
type WriteHooks interface {
	// OnInsert is called after an message is inserted into the store.
	OnInsert(context.Context, proto.Message)

	// OnUpdate is called after the entity is updated in the store.
	OnUpdate(ctx context.Context, existing, new proto.Message)

	// OnDelete is called after the entity is deleted from the store.
	OnDelete(context.Context, proto.Message)
}
