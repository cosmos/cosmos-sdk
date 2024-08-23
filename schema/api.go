package schema

// ModuleAPI represents a module or account's API.
type ModuleAPI struct {
	// Methods is the list of methods in the API.
	// It is a COMPATIBLE change to add new methods to an API.
	Methods []MethodType
}

// MethodType describes a method in the API.
type MethodType struct {
	// Name is the name of the method.
	Name string

	// InputFields is the list of input fields for the method.
	// It is an INCOMPATIBLE change to add, remove or update input fields to a method
	// and ObjectKind fields are not allowed.
	InputFields []Field

	// OutputFields is the list of output fields for the method.
	// It is a COMPATIBLE change to add new output fields to a method,
	// but existing output fields should not be removed or modified.
	// ObjectKind fields are allowed.
	OutputFields []Field
}
