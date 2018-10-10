package subspace

import (
	"reflect"
)

// TypeTable subspaces appropriate type for each parameter key
type TypeTable map[string]reflect.Type

// Constructs new table
func NewTypeTable(keytypes ...interface{}) (res TypeTable) {
	if len(keytypes)%2 != 0 {
		panic("odd number arguments in NewTypeTypeTable")
	}

	res = make(map[string]reflect.Type)

	for i := 0; i < len(keytypes); i += 2 {
		res = res.RegisterType(keytypes[i].([]byte), keytypes[i+1])
	}

	return
}

// Register single key-type pair
func (t TypeTable) RegisterType(key []byte, ty interface{}) TypeTable {
	keystr := string(key)
	if _, ok := t[keystr]; ok {
		panic("duplicate parameter key")
	}

	rty := reflect.TypeOf(ty)

	// Indirect rty if it is ptr
	if rty.Kind() == reflect.Ptr {
		rty = rty.Elem()
	}

	t[keystr] = rty

	return t
}

// Register multiple pairs from ParamSet
func (t TypeTable) RegisterParamSet(ps ParamSet) TypeTable {
	for _, kvp := range ps.KeyValuePairs() {
		t = t.RegisterType(kvp.Key, kvp.Value)
	}
	return t
}
