package container

import "reflect"

type Option interface {
	// TODO
	applyOption()
}

func Provide(constructors ...interface{}) Option {
	panic("TODO")
}

func ProvideWithScope(scope Scope, constructors ...interface{}) Option {
	panic("TODO")
}

func Options(opts ...Option) Option {
	panic("TODO")
}

func DefineGroupTypes(types ...reflect.Type) Option {
	panic("TODO")
}

func Error(err error) Option {
	panic("TODO")
}

type Scope struct {
	name string
}

func NewScope(name string) Scope {
	return Scope{name: name}
}

func (s Scope) Name() string {
	return s.name
}

type StructArgs struct{}

func Run(invoker interface{}, opts ...Option) error {
	panic("TODO")
}
