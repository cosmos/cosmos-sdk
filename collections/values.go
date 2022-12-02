package collections

var Uint64Value ValueEncoder[uint64] = uint64Value{}

type uint64Value struct{}

func (u uint64Value) Encode(value uint64) ([]byte, error) {
	return Uint64Key.Encode(value)
}

func (u uint64Value) Decode(b []byte) (uint64, error) {
	_, v, err := Uint64Key.Decode(b)
	return v, err
}

func (u uint64Value) Stringify(value uint64) string {
	return Uint64Key.Stringify(value)
}

func (u uint64Value) ValueType() string {
	return Uint64Key.KeyType()
}
