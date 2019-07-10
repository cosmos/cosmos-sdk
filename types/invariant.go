package types

import "fmt"

// An Invariant is a function which tests a particular invariant.
// If the invariant has been broken, it should return an error
// containing a descriptive message about what happened.
// The simulator will then halt and print the logs.
type Invariant func(ctx Context) (string, bool)

// Invariants defines a group of invariants
type Invariants []Invariant

// expected interface for registering invariants
type InvariantRegistry interface {
	RegisterRoute(moduleName, route string, invar Invariant)
}

// FormatInvariant returns a formatted invariant where module name and
// invariant name on the first line followed by the invariant message and
// its passing status.
func FormatInvariant(module, name, msg string, broken bool) (string, bool) {
	return fmt.Sprintf("%s: %s invariant\n%s\nInvariant Broken: %v\n",
		module, name, msg, broken), broken
}
