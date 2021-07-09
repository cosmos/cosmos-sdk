package container

import "reflect"

// Option is a functional option for a container
type Option interface {
	isOption()
}

// Provide creates a container option which registers the provided dependency
// injection constructors.
func Provide(constructors ...interface{}) Option {
	panic("TODO")
}

// ProvideWithScope creates a container option which registers the provided dependency
// injection constructors that are to be run in the provided scope.
func ProvideWithScope(scope Scope, constructors ...interface{}) Option {
	panic("TODO")
}

// AutoGroupTypes creates an option which registers the provided types as types which
// will automatically get grouped together. For example if
func AutoGroupTypes(types ...reflect.Type) Option {
	panic("TODO")
}

func OnePerScopeTypes(types ...reflect.Type) Option {
	panic("TODO")
}

// Error creates an option which causes the dependency injection failure to
// fail immediately.
func Error(err error) Option {
	panic("TODO")
}

// Options creates an option which bundles together many other options.
func Options(opts ...Option) Option {
	panic("TODO")
}
