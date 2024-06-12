package indexerbase

// ModuleSchema represents the logical schema of a module for purposes of indexing and querying.
type ModuleSchema struct {

	// Tables is a list of tables that are part of the schema for the module.
	Tables []Table
}
