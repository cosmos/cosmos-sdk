package schema

// APIDefinition is a public versioned descriptor of an API.
//
// Method handlers may accept different compatible encodings.
// Protobuf binary is one such encoding and protobuf mappings for each kind and type will be described (TODO).
//
// A "native" wire encoding is, however, already specified for each type and kind.
// While this encoding is intended to be used for storage purposes, it can also be used
// as for API method calls and has the advantage of being deterministic.
// If method handlers choose to use this "native" encoding, they should simply encode
// input and output parameters each field's value wire encoding.
type APIDefinition struct {
	// Name is the versioned name of the API.
	Name string

	// Methods is the list of methods in the API.
	// It is a COMPATIBLE change to add new methods to an API.
	// If a newer client tries to call a method that an older server does not recognize it,
	// an error will simply be returned.
	Methods []MethodType
}

// MethodType describes a method in the API.
type MethodType struct {
	// Name is the name of the method.
	Name string

	// InputFields is the list of input fields for the method.
	// It is an INCOMPATIBLE change to add, remove or update input fields to a method.
	// The addition of new fields introduces the possibility that a newer client
	// will send an incomprehensible message to an older server.
	// ObjectKind fields are NOT ALLOWED because it is possible to add new value fields
	// to an object type which would be an incompatible change for input fields.
	InputFields []Field

	// OutputFields is the list of output fields for the method.
	// It is a COMPATIBLE change to add new output fields to a method,
	// but existing output fields should not be removed or modified.
	// ObjectKind fields are ALLOWED.
	// If a newer client tries to call a method on an older server, the newer expected result output
	// fields will simply be populated with the default values for that field kind.
	OutputFields []Field
}
