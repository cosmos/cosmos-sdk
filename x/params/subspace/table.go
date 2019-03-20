package subspace

import (
	"reflect"
)

type attribute struct {
	ty reflect.Type
}

// KeyTable subspaces appropriate type for each parameter key
type KeyTable struct {
	m map[string]attribute
}

// Constructs new table
func NewKeyTable(keytypes ...interface{}) (res KeyTable) {
	if len(keytypes)%2 != 0 {
		panic("odd number arguments in NewTypeKeyTable")
	}

	res = KeyTable{
		m: make(map[string]attribute),
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
func (t KeyTable) RegisterType(key []byte, ty interface{}) KeyTable {
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

	t.m[keystr] = attribute{
		ty: rty,
	}

	return t
}

// Register multiple pairs from ParamSet
func (t KeyTable) RegisterParamSet(ps ParamSet) KeyTable {
	for _, kvp := range ps.ParamSetPairs() {
		t = t.RegisterType(kvp.Key, kvp.Value)
	}
	return t
}

func (t KeyTable) maxKeyLength() (res int) {
	for k := range t.m {
		l := len(k)
		if l > res {
			res = l
		}
	}
	return
}
