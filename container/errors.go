package container

import (
	"fmt"
	"reflect"

	reflect2 "github.com/cosmos/cosmos-sdk/container/reflect"
)

type duplicateConstructorError struct {
	loc reflect2.Location
	typ reflect.Type
}

func (d duplicateConstructorError) Error() string {
	return fmt.Sprintf("Duplicate constructor for type %v: %s", d.typ, d.loc)
}

type duplicateConstructorInScopeError struct {
	loc   reflect2.Location
	typ   reflect.Type
	scope Scope
}

func (d duplicateConstructorInScopeError) Error() string {
	return fmt.Sprintf("Duplicate constructor for one-per-scope type %v in scope %s: %s", d.typ, d.scope.Name(), d.loc)
}

var _, _ error = &duplicateConstructorError{}, &duplicateConstructorInScopeError{}
