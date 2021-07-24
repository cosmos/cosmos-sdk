package reflect

import (
	"reflect"
)

// Constructor defines a special constructor type that is defined by
// reflection. It should be passed as a value to the Provide function.
// Ex:
//   option.Provide(Constructor{ ... })
type Constructor struct {
	// In defines the in parameter types to Fn.
	In []Input

	// Out defines the out parameter types to Fn.
	Out []Output

	// Fn defines the constructor function.
	Fn func([]reflect.Value) ([]reflect.Value, error)

	// Location defines the source code location to be used for this constructor
	// in error messages.
	Location Location
}

type Input struct {
	Type     reflect.Type
	Optional bool
}

type Output struct {
	Type reflect.Type
}
