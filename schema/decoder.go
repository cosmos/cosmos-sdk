package schema

import "context"

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

	// ApplyUpdate is a function that applies an ObjectUpdate to the module's state for the given context.
	// If it is nil, the module doesn't support applying logical updates. If this function is provided
	// then it can be used as a genesis import path.
	ApplyUpdate ApplyUpdate
}

// KVDecoder is a function that decodes a key-value pair into an ObjectUpdate.
// If the KV-pair doesn't represent an object update, the function should return false
// as the second return value. Error should only be non-nil when the decoder expected
// to parse a valid update and was unable to.
type KVDecoder = func(key, value []byte) (ObjectUpdate, bool, error)

// ApplyUpdate is a function that applies an ObjectUpdate to the module's state for the given context.
type ApplyUpdate = func(context.Context, ObjectUpdate) error
