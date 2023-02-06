package collections

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
)

var (
	// Uint64Value implements a ValueCodec for uint64. It converts the uint64 to big endian bytes.
	// The JSON representation is the string format of uint64.
	Uint64Value ValueCodec[uint64] = uint64Value{}
	// StringValue implements a ValueCodec for string.
	StringValue ValueCodec[string] = stringValue{}
)

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
	return uint64EncodeJSON(value)
}

func (u uint64Value) DecodeJSON(b []byte) (uint64, error) {
	return uint64DecodeJSON(b)
}

func uint64EncodeJSON(value uint64) ([]byte, error) {
	str := `"` + strconv.FormatUint(value, 10) + `"`
	return []byte(str), nil
}

func uint64DecodeJSON(b []byte) (uint64, error) {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(str, 10, 64)
}

type stringValue struct{}

func (stringValue) Encode(value string) ([]byte, error) {
	return []byte(value), nil
}

func (stringValue) Decode(b []byte) (string, error) {
	return string(b), nil
}

func (stringValue) EncodeJSON(value string) ([]byte, error) {
	return json.Marshal(value)
}

func (stringValue) DecodeJSON(b []byte) (string, error) {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (stringValue) Stringify(value string) string {
	return value
}

func (stringValue) ValueType() string {
	return "string"
}
