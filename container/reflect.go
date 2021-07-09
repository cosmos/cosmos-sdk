package container

import "reflect"

type ReflectConstructor struct {
	In, Out  []reflect.Type
	Fn       func([]reflect.Value) []reflect.Value
	Location Location
}
