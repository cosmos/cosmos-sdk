package codegen

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// IsExportedType checks if the type is exported and not in an internal
// package.
func IsExportedType(typ reflect.Type) error {
	name := typ.Name()
	pkgPath := typ.PkgPath()
	if name != "" && pkgPath != "" {
		nameParts := strings.Split(name, ".")
		if unicode.IsLower([]rune(nameParts[len(nameParts)-1])[0]) {
			return errors.Errorf("type is not exported: %s", typ)
		}

		pkgParts := strings.Split(pkgPath, "/")
		if slices.Contains(pkgParts, "internal") {
			return errors.Errorf("type is in an internal package: %s", typ)
		}
	}

	switch typ.Kind() {
	case reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return nil

	case reflect.Array, reflect.Slice, reflect.Chan, reflect.Pointer:
		return IsExportedType(typ.Elem())

	case reflect.Func:
		numIn := typ.NumIn()
		for i := 0; i < numIn; i++ {
			err := IsExportedType(typ.In(i))
			if err != nil {
				return err
			}
		}

		numOut := typ.NumOut()
		for i := 0; i < numOut; i++ {
			err := IsExportedType(typ.Out(i))
			if err != nil {
				return err
			}
		}

		return nil

	case reflect.Map:
		err := IsExportedType(typ.Key())
		if err != nil {
			return err
		}
		return IsExportedType(typ.Elem())

	default:
		return nil
	}
}
