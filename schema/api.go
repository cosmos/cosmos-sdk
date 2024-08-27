package schema

// APIDefinition is a public versioned descriptor of an API.
//
// An APIDefinition can be used as a native descriptor of an API's encoding.
// The native binary encoding of API requests and responses is to encode the input and output
// fields using value binary encoding.
// The native JSON encoding would be to encode the fields as a JSON object, canonically
// sorted by field name with no extra whitespace.
// Thus, APIDefinitions have deterministic binary and JSON encodings.
//
// APIDefinitions have a strong definition of compatibility between different versions
// of the same API.
// It is compatible to add new methods to an API and to add new output fields
// to existing methods.
// It is incompatible to add new input fields to existing methods or to remove or modify
// existing input or output fields.
// Input fields cannot reference any types that can add new fields, such as ObjectKind fields.
// Adding new input fields to a method, directly or transitively, introduces the possibility that a newer client
// will send an incomprehensible message to an older server.
// The only safe ways that input field schemas can be extended are by adding
// new values to EnumType's and new cases to OneOfType's.
// Output fields can reference ObjectKind fields, which do allow for new value fields to be added.
// Object types are used to represent data in storage, so it is natural that object values
// can be returned as output fields of methods.
// If a newer client tries to call a method on an older server,
// any new output fields that the newer client knows about will simply be populated with their default values.
// If an older client tries to call a method on a newer server, the older client will simply ignore any new output fields.
//
// Existing protobuf APIs could also be mapped into APIDefinitions, and used in the following ways:
// - to produce, user-friendly deterministic JSON
// - to produce a deterministic binary encoding
// - to check for compatibility in a way that is more appropriate to blockchain applications
// - to use any code generators designed to support this spec as an alternative to protobuf
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
	// InputFields can only reference sealed StructTypes, either directly and transitively.
	InputFields []Field

	// OutputFields is the list of output fields for the method.
	// It is a COMPATIBLE change to add new output fields to a method,
	// but existing output fields should not be removed or modified.
	// OutputFields can reference any sealed or unsealed StructType, directly or transitively.
	// If a newer client tries to call a method on an older server, the newer expected result output
	// fields will simply be populated with the default values for that field kind.
	OutputFields []Field
}
