package module

import (
	"reflect"
)

type Registry struct {
	hmap        map[reflect.Type]interface{}
	handlerType reflect.Type
}

func NewRegistry(handlerType interface{}) *Registry {
	return &Registry{
		hmap:        map[reflect.Type]interface{}{},
		handlerType: reflect.TypeOf(handlerType),
	}
}

func (r *Registry) Register(constructor interface{}) {
	typ := reflect.TypeOf(constructor)
	if typ.Kind() != reflect.Func {
		panic("TODO")
	}

	if typ.NumIn() < 1 {
		panic("TODO")
	}

	configArg := typ.In(0)

	if _, ok := r.hmap[configArg]; ok {
		panic("TODO")
	}

	if typ.NumOut() < 1 {
		panic("TODO")
	}

	handlerOut := typ.Out(0)
	if !handlerOut.AssignableTo(r.handlerType) {
		panic("TODO")
	}

	r.hmap[configArg] = constructor
}

func (r *Registry) Resolve(configType interface{}) interface{} {
	return r.hmap[reflect.TypeOf(configType)]
}
