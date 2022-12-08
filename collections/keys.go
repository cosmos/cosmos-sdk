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

var errDecodeKeySize = errors.New("decode error, wrong byte key size")

type uint64Key struct{}

func (u uint64Key) Encode(buffer []byte, key uint64) (int, error) {
	binary.BigEndian.PutUint64(buffer, key)
	return 8, nil
}

func (u uint64Key) Decode(buffer []byte) (int, uint64, error) {
	if size := len(buffer); size < 8 {
		return 0, 0, fmt.Errorf("%w: wanted at least 8, got: %d", errDecodeKeySize, size)
	}
	return 8, binary.BigEndian.Uint64(buffer), nil
}

func (u uint64Key) Size(_ uint64) int { return 8 }

func (u uint64Key) Stringify(key uint64) string {
	return strconv.FormatUint(key, 10)
}

func (u uint64Key) KeyType() string {
	return "uint64"
}
