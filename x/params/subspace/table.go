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

func NewKeyTable(keyTypes ...interface{}) (res KeyTable) {
	if len(keyTypes)%2 != 0 {
		panic("odd number arguments in NewTypeKeyTable")
	}

	res = KeyTable{
		m: make(map[string]attribute),
	}

	for i := 0; i < len(keyTypes); i += 2 {
		res = res.RegisterType(NewParamSetPair(keyTypes[i].(Param), keyTypes[i+1]))
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

// RegisterType registers a single ParamSetPair (key-type pair) in a KeyTable.
func (t KeyTable) RegisterType(psp ParamSetPair) KeyTable {
	key := psp.Param.Key()

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

	rty := reflect.TypeOf(psp.Value)

	// Indirect rty if it is ptr
	if rty.Kind() == reflect.Ptr {
		rty = rty.Elem()
	}

	t.m[keystr] = attribute{
		ty: rty,
	}

	return t
}

// RegisterParamSet registers multiple ParamSetPairs from a ParamSet in a KeyTable.
func (t KeyTable) RegisterParamSet(ps ParamSet) KeyTable {
	for _, psp := range ps.ParamSetPairs() {
		t = t.RegisterType(psp)
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
