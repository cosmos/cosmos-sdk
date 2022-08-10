package depinject

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// isExportedType checks if the type is exported and not in an internal
// package.
func isExportedType(typ reflect.Type) error {
	name := typ.Name()
	pkgPath := typ.PkgPath()
	if name != "" && pkgPath != "" {
		if unicode.IsLower([]rune(name)[0]) {
			return errors.Errorf("type is not exported: %s", typ)
		}

		pkgParts := strings.Split(pkgPath, "/")
		if slices.Contains(pkgParts, "internal") {
			return errors.Errorf("type is in an internal package: %s", typ)
		}

		return nil
	}

	switch typ.Kind() {
	case reflect.Array, reflect.Slice, reflect.Chan, reflect.Pointer:
		return isExportedType(typ.Elem())

	case reflect.Func:
		numIn := typ.NumIn()
		for i := 0; i < numIn; i++ {
			err := isExportedType(typ.In(i))
			if err != nil {
				return err
			}
		}

		numOut := typ.NumOut()
		for i := 0; i < numOut; i++ {
			err := isExportedType(typ.Out(i))
			if err != nil {
				return err
			}
		}

		return nil

	case reflect.Map:
		err := isExportedType(typ.Key())
		if err != nil {
			return err
		}
		return isExportedType(typ.Elem())

	default:
		return nil
	}
}
