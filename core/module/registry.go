package module

import (
	"reflect"
)

type registry struct {
	hmap        map[reflect.Type]interface{}
	handlerType reflect.Type
}

func (r *registry) Register(constructor interface{}) {
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
