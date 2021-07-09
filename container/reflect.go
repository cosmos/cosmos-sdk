package container

import "reflect"

type ReflectConstructor struct {
	InType, OutTypes []reflect.Type
	Fn               func([]reflect.Value) []reflect.Value
	Location         Location
}
