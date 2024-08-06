package view

import "cosmossdk.io/schema"

// ObjectCollection is the interface for viewing the state of a collection of objects in a module
// represented by ObjectUpdate's for an ObjectType. ObjectUpdates must not include
// ValueUpdates in the Value field. When ValueUpdates are applied they must be
// converted to individual value or array format depending on the number of fields in
// the value. For collections which retain deletions, ObjectUpdate's with the Delete
// field set to true should be returned with the latest Value still intact.
type ObjectCollection interface {
	// ObjectType returns the object type for the collection.
	ObjectType() schema.ObjectType

	// GetObject returns the object update for the given key if it exists.
	GetObject(key any) (update schema.ObjectUpdate, found bool)

	// AllState iterates over the state of the collection by calling the given function with each item in
	// state represented as an object update.
	AllState(f func(schema.ObjectUpdate) bool)

	// Len returns the number of objects in the collection.
	Len() int
}
