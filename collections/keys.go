package collections

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
)

// Uint64Key can be used to encode uint64 keys.
// Encoding is big endian to retain ordering.
var Uint64Key KeyEncoder[uint64] = uint64Key{}

var errDecodeKeySize = errors.New("decode error, wrong byte key size")

type uint64Key struct{}

func (u uint64Key) Encode(key uint64) ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, key)
	return b, nil
}

func (u uint64Key) Decode(b []byte) (int, uint64, error) {
	if size := len(b); size != 8 {
		return 0, 0, fmt.Errorf("%w: wanted 8, got: %d", errDecodeKeySize, size)
	}
	return 8, binary.BigEndian.Uint64(b), nil
}

func (u uint64Key) Stringify(key uint64) string {
	return strconv.FormatUint(key, 10)
}

func (u uint64Key) KeyType() string {
	return "uint64"
}
