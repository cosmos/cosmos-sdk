package indexerbase

// ModuleSchema represents the logical schema of a module for purposes of indexing and querying.
type ModuleSchema struct {

	// ObjectTypes describe the types of objects that are part of the module's schema.
	ObjectTypes []ObjectType
}
