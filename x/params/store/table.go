package store

import (
	"reflect"
)

type Table struct {
	m      map[string]reflect.Type
	sealed bool
}

func NewTable(keytypes ...interface{}) (res Table) {
	if len(keytypes)%2 != 0 {
		panic("odd number arguments in NewTypeTable")
	}

	res = Table{
		m:      make(map[string]reflect.Type),
		sealed: false,
	}

	for i := 0; i < len(keytypes); i += 2 {
		res = res.RegisterType(keytypes[i].([]byte), keytypes[i+1])
	}

	return
}

func (t Table) RegisterType(key []byte, ty interface{}) Table {
	if t.sealed {
		panic("RegisterType() on sealed Table")
	}

	keystr := string(key)
	if _, ok := t.m[keystr]; ok {
		panic("duplicate parameter key")
	}

	rty := reflect.TypeOf(ty)

	// Indirect rty if it is ptr
	if rty.Kind() == reflect.Ptr {
		rty = rty.Elem()
	}

	t.m[keystr] = rty

	return t
}

func (t Table) RegisterParamStruct(ps ParamStruct) Table {
	for _, kvp := range ps.KeyValuePairs() {
		t = t.RegisterType(kvp.Key, kvp.Value)
	}
	return t
}
