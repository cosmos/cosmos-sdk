package collections

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
)

// Uint64Key can be used to encode uint64 keys.
// Encoding is big endian to retain ordering.
var Uint64Key KeyCodec[uint64] = uint64Key{}

var StringKey KeyCodec[string] = stringKey{}

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

func (uint64Key) Size(_ uint64) int { return 8 }

func (uint64Key) Stringify(key uint64) string {
	return strconv.FormatUint(key, 10)
}

func (uint64Key) KeyType() string {
	return "uint64"
}

type stringKey struct{}

func (s stringKey) Encode(buffer []byte, key string) (int, error) {
	return copy(buffer, key), nil
}

func (s stringKey) Decode(buffer []byte) (int, string, error) {
	return len(buffer), string(buffer), nil
}

func (s stringKey) Size(key string) int {
	return len(key)
}

func (s stringKey) Stringify(key string) string {
	return key
}

func (s stringKey) KeyType() string {
	return "string"
}
