package depinject

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

type ErrMultipleImplicitInterfaceBindings struct {
	error
	Interface reflect.Type
	Matches   []reflect.Type
}

func (err ErrMultipleImplicitInterfaceBindings) Error() string {
	return fmt.Sprintf("Multiple implementations found for interface %v: %v", err.Interface, err.Matches)
}

type ErrExplicitBindingNotFound struct {
	Preference preference
	Interface  reflect.Type
	error
}

func (err ErrExplicitBindingNotFound) Error() string {
	p := err.Preference
	if p.ModuleName != "" {
		return fmt.Sprintf("Given the explicit interface binding %s in module %s, a provider of type %s was not found.",
			p.Interface, p.ModuleName, p.Implementation)
	} else {
		return fmt.Sprintf("Given the explicit interface binding %s, a provider of type %s was not found.",
			p.Interface, p.Implementation)
	}

}

func duplicateDefinitionError(typ reflect.Type, duplicateLoc Location, existingLoc string) error {
	return errors.Errorf("duplicate provision of type %v by %s\n\talready provided by %s",
		typ, duplicateLoc, existingLoc)
}
