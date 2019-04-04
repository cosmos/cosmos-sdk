package types

// An Invariant is a function which tests a particular invariant.
// If the invariant has been broken, it should return an error
// containing a descriptive message about what happened.
// The simulator will then halt and print the logs.
type Invariant func(ctx Context) error

// group of Invarient
type Invariants []Invariant

// expected interface for routing invariants
type InvariantRouter interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}
