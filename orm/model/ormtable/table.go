package ormtable

import (
	"context"
	"encoding/json"
	"io"

	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
)

// View defines a read-only table.
//
// It exists as a separate interacted to support future scenarios where
// tables may be "supported" virtually to provide compatibility between
// systems, for instance to enable backwards compatibility when a major
// migration needs to be performed.
type View interface {
	UniqueIndex

	// GetIndex returns the index referenced by the provided fields if
	// one exists or nil. Note that some concrete indexes can be retrieved by
	// multiple lists of fields.
	GetIndex(fields string) Index

	// GetUniqueIndex returns the unique index referenced by the provided fields if
	// one exists or nil. Note that some concrete indexes can be retrieved by
	// multiple lists of fields.
	GetUniqueIndex(fields string) UniqueIndex

	// Indexes returns all the concrete indexes for the table.
	Indexes() []Index
}

// Table is an abstract interface around a concrete table. Table instances
// are stateless, with all state existing only in the store passed
// to table and index methods.
type Table interface {
	View

	ormkv.EntryCodec

	// Save saves the provided entry in the store either inserting it or
	// updating it if needed.
	//
	// If store implement the Hooks interface, the appropriate OnInsert or
	// OnUpdate hook method will be called.
	//
	// Save attempts to be atomic with respect to the underlying store,
	// meaning that either the full save operation is written or the store is
	// left unchanged, unless there is an error with the underlying store.
	Save(context context.Context, message proto.Message) error

	// Insert inserts the provided entry in the store and fails if there is
	// an unique key violation. See Save for more details on behavior.
	Insert(context context.Context, message proto.Message) error

	// Update updates the provided entry in the store and fails if an entry
	// with a matching primary key does not exist. See Save for more details
	// on behavior.
	Update(context context.Context, message proto.Message) error

	// Delete deletes the entry with the provided primary key from the store.
	//
	// If store implement the Hooks interface, the OnDelete hook method will
	// be called.
	//
	// Delete attempts to be atomic with respect to the underlying store,
	// meaning that either the full save operation is written or the store is
	// left unchanged, unless there is an error with the underlying store.
	Delete(context context.Context, message proto.Message) error

	// DefaultJSON returns default JSON that can be used as a template for
	// genesis files.
	//
	// For regular tables this an empty JSON array, but for singletons an
	// empty instance of the singleton is marshaled.
	DefaultJSON() json.RawMessage

	// ValidateJSON validates JSON streamed from the reader.
	ValidateJSON(io.Reader) error

	// ImportJSON imports JSON into the store, streaming one entry at a time.
	// Each table should be import from a separate JSON file to enable proper
	// streaming.
	//
	// Regular tables should be stored as an array of objects with each object
	// corresponding to a single record in the table.
	//
	// Auto-incrementing tables
	// can optionally have the last sequence value as the first element in the
	// array. If the last sequence value is provided, then each value of the
	// primary key in the file must be <= this last sequence value or omitted
	// entirely. If no last sequence value is provided, no entries should
	// contain the primary key as this will be auto-assigned.
	//
	// Singletons should define a single object and not an array.
	//
	// ImportJSON is not atomic with respect to the underlying store, meaning
	// that in the case of an error, some records may already have been
	// imported. It is assumed that ImportJSON is called in the context of some
	// larger transaction isolation.
	ImportJSON(context.Context, io.Reader) error

	// ExportJSON exports JSON in the format accepted by ImportJSON.
	// Auto-incrementing tables will export the last sequence number as the
	// first element in the JSON array.
	ExportJSON(context.Context, io.Writer) error

	// ID is the ID of this table within the schema of its FileDescriptor.
	ID() uint32
}
