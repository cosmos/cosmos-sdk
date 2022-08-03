package codegen

import (
	"fmt"
	"reflect"
)

// CheckIsExportedType checks if the type is exported and not in an internal
// package.
func CheckIsExportedType(typ reflect.Type) error {
	if name := typ.Name(); name != "" {
		// TODO generics
	}

	switch typ.Kind() {
	case reflect.Array, reflect.Slice, reflect.Chan, reflect.Pointer:
		return CheckIsExportedType(typ.Elem())

	case reflect.Func:
		return fmt.Errorf("TODO")

	case reflect.Map:
		err := CheckIsExportedType(typ.Key())
		if err != nil {
			return err
		}
		return CheckIsExportedType(typ.Elem())

	default:
		return nil
	}
}
