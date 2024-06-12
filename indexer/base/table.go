package indexerbase

// Table represents a table in the schema of a module.
type Table struct {
	// Name is the name of the table.
	Name string

	// KeyColumns is a list of columns that make up the primary key of the table.
	// It can be empty in which case indexers should assume that this table is
	// a singleton and ony has one value.
	KeyColumns []Column

	// ValueColumns is a list of columns that are not part of the primary key of the table.
	// It can be empty in the case where all columns are part of the primary key.
	ValueColumns []Column

	// RetainDeletions is a flag that indicates whether the indexer should retain
	// deleted rows in the database and flag them as deleted rather than actually
	// deleting the row. For many types of data in state, the data is deleted even
	// though it is still valid in order to save space. Indexers will want to have
	// the option of retaining such data and distinguishing from other "true" deletions.
	RetainDeletions bool
}
