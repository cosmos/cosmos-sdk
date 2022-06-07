package depinject

import (
	"fmt"
	"reflect"
)

// Preference defines a type binding preference to bind Interface to type Implementation when being provided as a
// dependency to the module with ModuleName.  If ModuleName is empty then the type binding is applied globally,
// not module-scoped.
type Preference struct {
	Interface      string
	Implementation string
	ModuleName     string
}

func fullyQualifiedTypeName(typ reflect.Type) string {
	return fmt.Sprintf("%s/%v", typ.PkgPath(), typ)
}

func findPreference(ps []Preference, typ reflect.Type, key *moduleKey) (Preference, bool) {
	if key != nil {
		for _, p := range ps {
			if p.Interface == fullyQualifiedTypeName(typ) && (key.name == p.ModuleName) {
				return p, true
			}
		}
	}

	for _, p := range ps {
		if p.Interface == fullyQualifiedTypeName(typ) && p.ModuleName == "" {
			return p, true
		}
	}

	return Preference{}, false
}
