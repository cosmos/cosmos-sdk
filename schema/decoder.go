package schema

// HasModuleCodec is an interface that modules can implement to provide a ModuleCodec.
// Usually these modules would also implement appmodule.AppModule, but that is not included
// to keep this package free of any dependencies.
type HasModuleCodec interface {
	// ModuleCodec returns a ModuleCodec for the module.
	ModuleCodec() (ModuleCodec, error)
}

// ModuleCodec is a struct that contains the schema and a KVDecoder for a module.
type ModuleCodec struct {
	// Schema is the schema for the module. It is required.
	Schema ModuleSchema

	// KVDecoder is a function that decodes a key-value pair into an ObjectUpdate.
	// If it is nil, the module doesn't support state decoding directly.
	KVDecoder KVDecoder
}

// KVDecoder is a function that decodes a key-value pair into an ObjectUpdate.
// If the KV-pair doesn't represent an object update, the function should return false
// as the second return value. Error should only be non-nil when the decoder expected
// to parse a valid update and was unable to.
type KVDecoder = func(key, value []byte) (ObjectUpdate, bool, error)
