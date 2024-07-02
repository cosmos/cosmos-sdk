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

// KVDecoder is a function that decodes a key-value pair into one or more ObjectUpdate's.
// If the KV-pair doesn't represent object updates, the function should return nil as the first
// and no error. The error result  should only be non-nil when the decoder expected
// to parse a valid update and was unable to. In the case of an error, the decoder may return
// a non-nil value for the first return value, which can indicate which parts of the update
// were decodable to aid debugging.
type KVDecoder = func(KVPairUpdate) ([]ObjectUpdate, error)

type KVPairUpdate struct {
	// Key is the key of the key-value pair.
	Key []byte

	// Value is the value of the key-value pair. It should be ignored when Delete is true.
	Value []byte

	// Delete is a flag that indicates that the key-value pair was deleted. If it is false,
	// then it is assumed that this has been a set operation.
	Delete bool
}

// ApplyUpdate is a function that applies an ObjectUpdate to the module's state for the given context.
type ApplyUpdate = func(context.Context, ObjectUpdate) error
