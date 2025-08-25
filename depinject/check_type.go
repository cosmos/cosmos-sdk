package depinject

import (
	"errors"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	errTypeMustBeExported  = errors.New("type must be exported")
	errTypeFromInternalPkg = errors.New("type must not come from an internal package")
)

// isExportedType checks if the type is exported and not in an internal
// package. NOTE: generic type parameters are not checked because this
// would involve complex parsing of type names (there is no reflect API for
// generic type parameters). Parsing of these parameters should be possible
// if someone chooses to do it in the future, but care should be taken to
// be exhaustive and cover all cases like pointers, map's, chan's, etc. which
// means you actually need a real parser and not just a regex.
func isExportedType(typ reflect.Type) error {
	name := typ.Name()
	pkgPath := typ.PkgPath()
	if name != "" && pkgPath != "" {
		if r, _ := utf8.DecodeRuneInString(name); unicode.IsLower(r) {
			return errTypeMustBeExported
		}

		if strings.Contains(pkgPath, "/internal/") ||
			strings.HasSuffix(pkgPath, "/internal") ||
			strings.HasPrefix(pkgPath, "internal/") {
			return errTypeFromInternalPkg
		}

		return nil
	}

	switch typ.Kind() {
	case reflect.Array, reflect.Slice, reflect.Chan, reflect.Pointer:
		return isExportedType(typ.Elem())

	case reflect.Func:
		for i := 0; i < typ.NumIn(); i++ {
			if err := isExportedType(typ.In(i)); err != nil {
				return err
			}
		}

		for i := 0; i < typ.NumOut(); i++ {
			if err := isExportedType(typ.Out(i)); err != nil {
				return err
			}
		}

		return nil
	case reflect.Map:
		if err := isExportedType(typ.Key()); err != nil {
			return err
		}
		return isExportedType(typ.Elem())

	default:
		// all the remaining types are builtin, non-composite types (like integers), so they are fine to use
		return nil
	}
}
