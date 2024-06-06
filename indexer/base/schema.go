package indexerbase

type Schema struct {
	Tables []Table
}

type Table struct {
	Name         string
	KeyColumns   []Column
	ValueColumns []Column

	// RetainDeletions is a flag that indicates whether the indexer should retain
	// deleted rows in the database and flag them as deleted rather than actually
	// deleting the row. For many types of data in state, the data is deleted even
	// though it is still valid in order to save space. Indexers will want to have
	// the option of retaining such data and distinguishing from other "true" deletions.
	RetainDeletions bool
}

type Column struct {
	Name           string
	Type           Type
	Nullable       bool
	AddressPrefix  string
	EnumDefinition EnumDefinition
}

type EnumDefinition struct {
	Name   string
	Values []string
}
