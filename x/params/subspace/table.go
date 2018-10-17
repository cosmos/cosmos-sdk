package subspace

import (
	"reflect"
)

// TypeTable subspaces appropriate type for each parameter key
type TypeTable struct {
	m map[string]reflect.Type
}

// Constructs new table
func NewTypeTable(keytypes ...interface{}) (res TypeTable) {
	if len(keytypes)%2 != 0 {
		panic("odd number arguments in NewTypeTypeTable")
	}

	res = TypeTable{
		m: make(map[string]reflect.Type),
	}

	for i := 0; i < len(keytypes); i += 2 {
		res = res.RegisterType(keytypes[i].([]byte), keytypes[i+1])
	}

	return
}

func isAlphaNumeric(key []byte) bool {
	for _, b := range key {
		if !((48 <= b && b <= 57) || // numeric
			(65 <= b && b <= 90) || // upper case
			(97 <= b && b <= 122)) { // lower case
			return false
		}
	}
	return true
}

// Register single key-type pair
func (t TypeTable) RegisterType(key []byte, ty interface{}) TypeTable {
	if len(key) == 0 {
		panic("cannot register empty key")
	}
	if !isAlphaNumeric(key) {
		panic("non alphanumeric parameter key")
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

// Register multiple pairs from ParamSet
func (t TypeTable) RegisterParamSet(ps ParamSet) TypeTable {
	for _, kvp := range ps.KeyValuePairs() {
		t = t.RegisterType(kvp.Key, kvp.Value)
	}
	return t
}

func (t TypeTable) maxKeyLength() (res int) {
	for k := range t.m {
		l := len(k)
		if l > res {
			res = l
		}
	}
	return
}
