package schema

// APIDescriptor is a public versioned descriptor of an API.
//
// An APIDescriptor can be used as a native descriptor of an API's encoding.
// The native binary encoding of API requests and responses is to encode the input and output
// fields using value binary encoding.
// The native JSON encoding would be to encode the fields as a JSON object, canonically
// sorted by field name with no extra whitespace.
// Thus, APIDefinitions have deterministic binary and JSON encodings.
//
// APIDefinitions have a strong definition of compatibility between different versions
// of the same API.
// It is an INCOMPATIBLE change to add new input fields to existing methods or to remove or modify
// existing input or output fields.
// Input fields also cannot reference any unsealed structs, directly or transitively,
// because these types allow adding new fields.
// Adding new input fields to a method introduces the possibility that a newer client
// will send an incomprehensible message to an older server.
// The only safe ways that input field schemas can be extended are by adding
// new values to EnumType's and new cases to OneOfType's.
// It is a COMPATIBLE change to add new methods to an API and to add new output fields
// to existing methods.
// Output fields can reference any sealed or unsealed StructType, directly or transitively.
//
// Existing protobuf APIs could also be mapped into APIDefinitions, and used in the following ways:
// - to produce, user-friendly deterministic JSON
// - to produce a deterministic binary encoding
// - to check for compatibility in a way that is more appropriate to blockchain applications
// - to use any code generators designed to support this spec as an alternative to protobuf
// Also, a standardized way of serializing schema types as protobuf could be defined which
// maps to the original protobuf encoding, so that schemas can be used as an interop
// layer between different less expressive encoding systems.
//
// Existing EVM contract APIs expressed in Solidity could be mapped into APIDefinitions, and
// a mapping of all schema values to ABI encoding could be defined which preserves the
// original ABI encoding.
//
// In this way, we can define an interop layer between contracts in the EVM world,
// SDK modules accepting protobuf types, and any API using this schema system natively.
type APIDescriptor struct {
	// Name is the versioned name of the API.
	Name string

	// Methods is the list of methods in the API.
	// It is a COMPATIBLE change to add new methods to an API.
	// If a newer client tries to call a method that an older server does not recognize it,
	// an error will simply be returned.
	Methods []MethodDescriptor
}

// MethodDescriptor describes a method in the API.
type MethodDescriptor struct {
	// Name is the name of the method.
	Name string

	// InputFields is the list of input fields for the method.
	//
	// It is an INCOMPATIBLE change to add, remove or update input fields to a method.
	// The addition of new fields introduces the possibility that a newer client
	// will send an incomprehensible message to an older server.
	// InputFields can only reference sealed StructTypes, either directly and transitively.
	//
	// As a special case to represent protobuf service definitions, there can be a single
	// unnamed struct input field that code generators can choose to either reference
	// as a named struct or to expand inline as function arguments.
	InputFields []Field

	// OutputFields is the list of output fields for the method.
	//
	// It is a COMPATIBLE change to add new output fields to a method,
	// but existing output fields should not be removed or modified.
	// OutputFields can reference any sealed or unsealed StructType, directly or transitively.
	// If a newer client tries to call a method on an older server, the newer expected result output
	// fields will simply be populated with the default values for that field kind.
	//
	// As a special case to represent protobuf service definitions, there can be a single
	// unnamed struct output field.
	// In this case, adding new output fields is an INCOMPATIBLE change (because protobuf service definitions
	// don't allow this), but new fields can be added to the referenced struct if it is unsealed.
	OutputFields []Field

	// Volatility is the volatility of the method.
	Volatility Volatility
}

// Volatility is the volatility of a method.
type Volatility int

const (
	// PureVolatility indicates that the method can neither read nor write state.
	PureVolatility Volatility = iota
	// ReadonlyVolatility indicates that the method can read state but not write state.
	ReadonlyVolatility

	// VolatileVolatility indicates that the method can read and write state.
	VolatileVolatility
)
