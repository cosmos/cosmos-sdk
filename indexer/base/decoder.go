package indexerbase

// DecodableModule is an interface that modules can implement to provide a ModuleDecoder.
// Usually these modules would also implement appmodule.AppModule, but that is not included
// to keep this package free of any dependencies.
type DecodableModule interface {
	// ModuleDecoder returns a ModuleDecoder for the module.
	ModuleDecoder() (ModuleDecoder, error)
}

// ModuleDecoder is a struct that contains the schema and a KVDecoder for a module.
type ModuleDecoder struct {
	// Schema is the schema for the module.
	Schema ModuleSchema

	// KVDecoder is a function that decodes a key-value pair into an ObjectUpdate.
	// If modules pass logical updates directly to the engine and don't require logical decoding of raw bytes,
	// then this function should be nil.
	KVDecoder KVDecoder
}

// KVDecoder is a function that decodes a key-value pair into an ObjectUpdate.
// If the KV-pair doesn't represent an object update, the function should return false
// as the second return value. Error should only be non-nil when the decoder expected
// to parse a valid update and was unable to.
type KVDecoder = func(key, value []byte) (ObjectUpdate, bool, error)
