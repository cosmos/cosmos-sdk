package indexerbase

// EntityUpdate represents an update operation on an entity in the schema.
type EntityUpdate struct {
	// TableName is the name of the table that the entity belongs to in the schema.
	TableName string

	// Key returns the value of the primary key of the entity and must conform to these constraints with respect
	// that the schema that is defined for the entity:
	// - if key represents a single column, then the value must be valid for the first column in that
	// 	column list. For instance, if there is one column in the key of type String, then the value must be of
	//  type string
	// - if key represents multiple columns, then the value must be a slice of values where each value is valid
	//  for the corresponding column in the column list. For instance, if there are two columns in the key of
	//  type String, String, then the value must be a slice of two strings.
	// If the key has no columns, meaning that this is a singleton entity, then this value is ignored and can be nil.
	Key any

	// Value returns the non-primary key columns of the entity and can either conform to the same constraints
	// as EntityUpdate.Key or it may be and instance of ValueUpdates. ValueUpdates can be used as a performance
	// If this is a delete operation, then this value is ignored and can be nil.
	Value any

	Delete bool
}

// ValueUpdates is an interface that represents the value columns of an entity update. Columns that
// were not updated may be excluded from the update. Consumers should be aware that implementations
// may not filter out columns that were unchanged. However, if a column is omitted from the update
// it should be considered unchanged.
type ValueUpdates interface {

	// Iterate iterates over the columns and values in the entity update. The function should return
	// true to continue iteration or false to stop iteration. Each column value should conform
	// to the requirements of that column's type in the schema. Iterate returns an error if
	// it was unable to decode the values properly (which could be the case in lazy evaluation).
	Iterate(func(col string, value any) bool) error
}
