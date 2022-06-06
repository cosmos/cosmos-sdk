package depinject

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

type ErrMultipleImplicitInterfaceBindings struct {
	Interface reflect.Type
	Err       error
}

func (err ErrMultipleImplicitInterfaceBindings) Error() string {
	return fmt.Sprintf("Multiple implementations found for interface %v", err.Interface)
}

func duplicateDefinitionError(typ reflect.Type, duplicateLoc Location, existingLoc string) error {
	return errors.Errorf("duplicate provision of type %v by %s\n\talready provided by %s",
		typ, duplicateLoc, existingLoc)
}
