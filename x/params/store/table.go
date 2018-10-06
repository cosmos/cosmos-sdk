package store

import (
	"reflect"
)

// Table stores appropriate type for each parameter key
type Table map[string]reflect.Type

// Constructs new table
func NewTable(keytypes ...interface{}) (res Table) {
	if len(keytypes)%2 != 0 {
		panic("odd number arguments in NewTypeTable")
	}

	res = make(map[string]reflect.Type)

	for i := 0; i < len(keytypes); i += 2 {
		res = res.RegisterType(keytypes[i].([]byte), keytypes[i+1])
	}

	return
}

// Register single key-type pair
func (t Table) RegisterType(key []byte, ty interface{}) Table {
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

// Register multiple pairs from ParamStruct
func (t Table) RegisterParamStruct(ps ParamStruct) Table {
	for _, kvp := range ps.KeyValuePairs() {
		t = t.RegisterType(kvp.Key, kvp.Value)
	}
	return t
}
