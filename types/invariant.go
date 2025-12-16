package types

import "fmt"

// An Invariant is a function which tests a particular invariant.
// The invariant returns a descriptive message about what happened
// and a boolean indicating whether the invariant has been broken.
// The simulator will then halt and print the logs.
//
// Deprecated: the Invariant type is deprecated and will be removed once x/crisis is removed.
type Invariant func(ctx Context) (string, bool)

// Invariants defines a group of invariants
//
// Deprecated: the Invariants type is deprecated and will be removed once x/crisis is removed.
type Invariants []Invariant

// InvariantRegistry is the expected interface for registering invariants
//
// Deprecated: the InvariantRegistry type is deprecated and will be removed once x/crisis is removed.
type InvariantRegistry interface {
	RegisterRoute(moduleName, route string, invar Invariant)
}

// FormatInvariant returns a standardized invariant message.
//
// Deprecated: the FormatInvariant type is deprecated and will be removed once x/crisis is removed.
func FormatInvariant(module, name, msg string) string {
	return fmt.Sprintf("%s: %s invariant\n%s\n", module, name, msg)
}
