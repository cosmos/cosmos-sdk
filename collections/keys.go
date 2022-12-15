package collections

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var (
	// Uint64Key can be used to encode uint64 keys.
	// Encoding is big endian to retain ordering.
	Uint64Key KeyCodec[uint64] = uint64Key{}
	// StringKey can be used to encode string keys.
	// The encoding just converts the string to bytes.
	StringKey KeyCodec[string] = stringKey{}
)
var errDecodeKeySize = errors.New("decode error, wrong byte key size")

type uint64Key struct{}

func (u uint64Key) Encode(buffer []byte, key uint64) (int, error) {
	binary.BigEndian.PutUint64(buffer, key)
	return 8, nil
}

func (uint64Key) Decode(buffer []byte) (int, uint64, error) {
	if size := len(buffer); size < 8 {
		return 0, 0, fmt.Errorf("%w: wanted at least 8, got: %d", errDecodeKeySize, size)
	}
	return 8, binary.BigEndian.Uint64(buffer), nil
}

func (uint64Key) EncodeJSON(value uint64) ([]byte, error) {
	return uint64EncodeJSON(value)
}

func (uint64Key) DecodeJSON(b []byte) (uint64, error) {
	return uint64DecodeJSON(b)
}

func (uint64Key) Size(_ uint64) int { return 8 }

func (uint64Key) Stringify(key uint64) string {
	return strconv.FormatUint(key, 10)
}

func (uint64Key) KeyType() string {
	return "uint64"
}

type stringKey struct{}

func (stringKey) Encode(buffer []byte, key string) (int, error) {
	return copy(buffer, key), nil
}

func (stringKey) Decode(buffer []byte) (int, string, error) {
	return len(buffer), string(buffer), nil
}

func (stringKey) EncodeJSON(value string) ([]byte, error) {
	return json.Marshal(value)
}

func (stringKey) DecodeJSON(b []byte) (string, error) {
	var value string
	err := json.Unmarshal(b, &value)
	return value, err
}

func (stringKey) Size(key string) int {
	return len(key)
}

func (stringKey) Stringify(key string) string {
	return key
}

func (stringKey) KeyType() string {
	return "string"
}
