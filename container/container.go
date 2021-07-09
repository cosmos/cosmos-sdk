package container

import "reflect"

type container struct {
	err                  error
	autoGroupTypes       []reflect.Type
	autoNameByScopeTypes []reflect.Type

	providers      map[reflect.Type]*node
	scopeProviders map[reflect.Type]*scopeNode
	nodes          []*node
	scopeNodes     []*scopeNode

	values          map[key]secureValue
	scopedValues    map[Scope]map[key]reflect.Value
	securityContext func(scope Scope, tag string) error
}

type input struct {
	key
	Optional bool
}

type Output struct {
	key
	SecurityChecker securityChecker
}

type key struct {
	Type reflect.Type
}

type node struct {
	Provider
	called bool
	values []reflect.Value
	err    error
}

// Provider is a general dependency provider. Its scope parameter is used
// to receive scoped dependencies and gain access to general dependencies within
// its security policy. Access to dependencies provided by this provider can optionally
// be restricted to certain scopes based on SecurityCheckers.
type Provider struct {
	// Constructor provides the dependencies
	Constructor func(deps []reflect.Value, scope Scope) ([]reflect.Value, error)

	// Needs are the keys for dependencies the constructor needs
	Needs []input

	// Needs are the keys for dependencies the constructor provides
	Provides []Output

	// Scope is the scope within which the constructor runs
	Scope Scope

	IsScopeProvider bool
}

type scopeNode struct {
	Provider
	calledForScope map[Scope]bool
	valuesForScope map[Scope][]reflect.Value
	errsForScope   map[Scope]error
}

// ScopeProvider provides scoped dependencies. Its constructor function will provide
// dependencies specific to the scope parameter. Instead of providing general dependencies
// with restricted access based on security checkers, ScopeProvider provides potentially different
// dependency instances to different scopes. It is assumed that a scoped provider
// can provide a dependency for any valid scope passed to it, although it can return an error
// to deny access.
type ScopeProvider struct {

	// Constructor provides dependencies for the provided scope
	Constructor func(scope Scope, deps []reflect.Value) ([]reflect.Value, error)

	// Needs are the keys for dependencies the constructor needs
	Needs []input

	// Needs are the keys for dependencies the constructor provides
	Provides []key

	// Scope is the scope within which the constructor runs, if it is left empty,
	// the constructor runs in the scope it was called with (this only applies to ScopeProvider).
	Scope Scope
}

type secureValue struct {
	value           reflect.Value
	securityChecker securityChecker
}

type securityChecker func(scope Scope) error
