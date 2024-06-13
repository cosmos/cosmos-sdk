package indexerbase

// ModuleSchema represents the logical schema of a module for purposes of indexing and querying.
type ModuleSchema struct {

	// Objects is a list of objects that are part of the schema for the module.
	Objects []ObjectDescriptor
}
