package indexerbase

// ObjectType describes an object type a module schema.
type ObjectType struct {
	// KeyFields is a list of fields that make up the primary key of the object.
	// It can be empty in which case indexers should assume that this object is
	// a singleton and ony has one value.
	KeyFields []Field

	// ValueFields is a list of fields that are not part of the primary key of the object.
	// It can be empty in the case where all fields are part of the primary key.
	ValueFields []Field

	// RetainDeletions is a flag that indicates whether the indexer should retain
	// deleted rows in the database and flag them as deleted rather than actually
	// deleting the row. For many types of data in state, the data is deleted even
	// though it is still valid in order to save space. Indexers will want to have
	// the option of retaining such data and distinguishing from other "true" deletions.
	RetainDeletions bool
}
