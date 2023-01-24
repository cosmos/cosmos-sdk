package math

import "cosmossdk.io/collections"

// IntValue represents a collections.ValueCodec to work with Int.
var IntValue collections.ValueCodec[Int] = intValueCodec{}

type intValueCodec struct{}

func (i intValueCodec) Encode(value Int) ([]byte, error) {
	return value.Marshal()
}

func (i intValueCodec) Decode(b []byte) (Int, error) {
	v := new(Int)
	err := v.Unmarshal(b)
	if err != nil {
		return Int{}, err
	}
	return *v, nil
}

func (i intValueCodec) EncodeJSON(value Int) ([]byte, error) {
	return value.MarshalJSON()
}

func (i intValueCodec) DecodeJSON(b []byte) (Int, error) {
	v := new(Int)
	err := v.UnmarshalJSON(b)
	if err != nil {
		return Int{}, err
	}
	return *v, nil
}

func (i intValueCodec) Stringify(value Int) string {
	return value.String()
}

func (i intValueCodec) ValueType() string {
	return "math.Int"
}
