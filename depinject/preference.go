package depinject

import (
	"fmt"
	"reflect"
)

// preference defines a type binding preference to bind interfaceName to type implTypeName when being provided as a
// dependency to the module identified by moduleKey.  If moduleKey is nil then the type binding is applied globally,
// not module-scoped.
type preference struct {
	interfaceName string
	implTypeName  string
	moduleKey     *moduleKey
	resolver      resolver
}

func fullyQualifiedTypeName(typ reflect.Type) string {
	pkgType := typ
	if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map || typ.Kind() == reflect.Array {
		pkgType = typ.Elem()
	}
	return fmt.Sprintf("%s/%v", pkgType.PkgPath(), typ)
}

func preferenceKeyFromTypeName(typeName string, key *moduleKey) string {
	if key == nil {
		return fmt.Sprintf("%s;", typeName)
	}
	return fmt.Sprintf("%s;%s", typeName, key.name)
}

func preferenceKeyFromType(typ reflect.Type, key *moduleKey) string {
	return preferenceKeyFromTypeName(fullyQualifiedTypeName(typ), key)
}
