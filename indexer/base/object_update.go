package indexerbase

// ObjectUpdate represents an update operation on an object in a module's state.
type ObjectUpdate struct {

	// ObjectName is the name of the object type in the module's schema.
	ObjectName string

	// Key returns the value of the primary key of the object and must conform to these constraints with respect
	// that the schema that is defined for the object:
	// - if key represents a single field, then the value must be valid for the first field in that
	// 	field list. For instance, if there is one field in the key of type String, then the value must be of
	//  type string
	// - if key represents multiple fields, then the value must be a slice of values where each value is valid
	//  for the corresponding field in the field list. For instance, if there are two fields in the key of
	//  type String, String, then the value must be a slice of two strings.
	// If the key has no fields, meaning that this is a singleton object, then this value is ignored and can be nil.
	Key any

	// Value returns the non-primary key fields of the object and can either conform to the same constraints
	// as ObjectUpdate.Key or it may be and instance of ValueUpdates. ValueUpdates can be used as a performance
	// optimization to avoid copying the values of the object into the update and/or to omit unchanged fields.
	// If this is a delete operation, then this value is ignored and can be nil.
	Value any

	// Delete is a flag that indicates whether this update is a delete operation. If true, then the Value field
	// is ignored and can be nil.
	Delete bool
}

// ValueUpdates is an interface that represents the value fields of an object update. fields that
// were not updated may be excluded from the update. Consumers should be aware that implementations
// may not filter out fields that were unchanged. However, if a field is omitted from the update
// it should be considered unchanged.
type ValueUpdates interface {

	// Iterate iterates over the fields and values in the object update. The function should return
	// true to continue iteration or false to stop iteration. Each field value should conform
	// to the requirements of that field's type in the schema. Iterate returns an error if
	// it was unable to decode the values properly (which could be the case in lazy evaluation).
	Iterate(func(col string, value any) bool) error
}
