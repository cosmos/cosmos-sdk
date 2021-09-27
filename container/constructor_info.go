package container

import "reflect"

// ConstructorInfo defines a special constructor type that is defined by
// reflection. It should be passed as a value to the Provide function.
// Ex:
//   option.Provide(ConstructorInfo{ ... })
type ConstructorInfo struct {
	// In defines the in parameter types to Fn.
	In []reflect.Type

	// Out defines the out parameter types to Fn.
	Out []reflect.Type

	// Fn defines the constructor function.
	Fn func([]reflect.Value) []reflect.Value

	// Location defines the source code location to be used for this constructor
	// in error messages.
	Location Location
}
