package container

import (
	"fmt"
)

// Location describes the source code location of a dependency injection
// constructor.
type Location interface {
	isLocation()
	fmt.Stringer
	fmt.Formatter
}

// LocationFromPC builds a Location from a function program counter location,
// such as that returned by reflect.Value.Pointer() or runtime.Caller().
func LocationFromPC(pc uintptr) Location {
	panic("TODO")
}
