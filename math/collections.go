package math

type IntValueCodec struct{}

func (i IntValueCodec) Encode(value Int) ([]byte, error) {
	return value.Marshal()
}

func (i IntValueCodec) Decode(b []byte) (Int, error) {
	v := new(Int)
	err := v.Unmarshal(b)
	if err != nil {
		return Int{}, err
	}
	return *v, nil
}

func (i IntValueCodec) EncodeJSON(value Int) ([]byte, error) {
	return value.MarshalJSON()
}

func (i IntValueCodec) DecodeJSON(b []byte) (Int, error) {
	v := new(Int)
	err := v.UnmarshalJSON(b)
	if err != nil {
		return Int{}, err
	}
	return *v, nil
}

func (i IntValueCodec) Stringify(value Int) string {
	return value.String()
}

func (i IntValueCodec) ValueType() string {
	return "math.Int"
}
