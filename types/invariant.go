package types

// An Invariant is a function which tests a particular invariant.
// If the invariant has been broken, it should return an error
// containing a descriptive message about what happened.
// The simulator will then halt and print the logs.
type Invariant func(ctx Context) error

// Invariants defines a group of invariants
type Invariants []Invariant

// expected interface for registering invariants
type InvariantRegistry interface {
	RegisterRoute(moduleName, route string, invar Invariant)
}
