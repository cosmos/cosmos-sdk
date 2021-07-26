package container

import (
	"reflect"

	"github.com/pkg/errors"
)

func duplicateConstructorError(typ reflect.Type, duplicateLoc Location, existingLoc string) error {
	return errors.Errorf("duplicate constructor for type %v: %s\n\talready defined by %s",
		typ, duplicateLoc, existingLoc)
}

func dependencyCycleError(failingLocation Location, failingType reflect.Type, callStack []Location) error {
	panic("TODO")
}

func cantResolveError(typ reflect.Type, resolveStack []resolveFrame) error {
	panic("TODO")
}
