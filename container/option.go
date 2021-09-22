package container

import "reflect"

// Option is a functional option for a container.
type Option interface {
	isOption()
}

// Provide creates a container option which registers the provided dependency
// injection constructors. Each constructor will be called at most once with the
// exception of scoped constructors which are called at most once per scope
// (see Scope).
func Provide(constructors ...interface{}) Option {
	panic("TODO")
}

// ProvideWithScope creates a container option which registers the provided dependency
// injection constructors that are to be run in the provided scope. Each constructor
// will be called at most once.
func ProvideWithScope(scope Scope, constructors ...interface{}) Option {
	panic("TODO")
}

// AutoGroupTypes creates an option which registers the provided types as types which
// will automatically get grouped together. For a given type T, T and []T can
// be declared as output parameters for constructors as many times within the container
// as desired. All of the provided values for T can be retrieved by declaring an
// []T input parameter.
func AutoGroupTypes(types ...reflect.Type) Option {
	panic("TODO")
}

// OnePerScopeTypes creates an option which registers the provided types as types which
// can have up to one value per scope. All of the values for a one-per-scope type T
// and their respective scopes, can be retrieved by declaring an input parameter map[Scope]T.
func OnePerScopeTypes(types ...reflect.Type) Option {
	panic("TODO")
}

// Error creates an option which causes the dependency injection container to
// fail immediately.
func Error(err error) Option {
	panic("TODO")
}

// Options creates an option which bundles together other options.
func Options(opts ...Option) Option {
	panic("TODO")
}
