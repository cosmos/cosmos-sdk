package container

import (
	"reflect"

	"github.com/pkg/errors"
)

func duplicateDefinitionError(typ reflect.Type, duplicateLoc Location, existingLoc string) error {
	return errors.Errorf("duplicate provision of type %v by %s\n\talready provided by %s",
		typ, duplicateLoc, existingLoc)
}

func dependencyCycleError(failingLocation Location, failingType reflect.Type, callStack []Location) error {
	panic("TODO")
}

func cantResolveError(typ reflect.Type, resolveStack []resolveFrame) error {
	panic("TODO")
}
