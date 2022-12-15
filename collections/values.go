package collections

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

var Uint64Value ValueCodec[uint64] = uint64Value{}

type uint64Value struct{}

func (u uint64Value) Encode(value uint64) ([]byte, error) {
	return binary.BigEndian.AppendUint64(make([]byte, 0, 8), value), nil
}

func (u uint64Value) Decode(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, fmt.Errorf("%w: uint64 value size invalid, want: 8, got: %d", ErrEncoding, len(b))
	}
	return binary.BigEndian.Uint64(b), nil
}

func (u uint64Value) Stringify(value uint64) string {
	return Uint64Key.Stringify(value)
}

func (u uint64Value) ValueType() string {
	return Uint64Key.KeyType()
}

func (u uint64Value) EncodeJSON(value uint64) ([]byte, error) {
	return json.Marshal(value)
}

func (u uint64Value) DecodeJSON(b []byte) (uint64, error) {
	var value uint64
	err := json.Unmarshal(b, &value)
	return value, err
}
