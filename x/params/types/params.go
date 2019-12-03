package types

// Param defines a parameter that any module may store or retrieve in a Subspace.
// A Param is primarily composed of a key and a validation function that returns
// an error if the Param value is considered invalid.
type Param interface {
	Key() []byte
	Validate() error
}
