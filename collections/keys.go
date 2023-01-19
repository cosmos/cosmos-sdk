package collections

import (
	"bytes"
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

// errDecodeKeySize is a sentinel error.
var errDecodeKeySize = errors.New("decode error, wrong byte key size")

// StringDelimiter defines the delimiter of a string key when used in non-terminal encodings.
const StringDelimiter uint8 = 0x0

type uint64Key struct{}

func (uint64Key) Encode(buffer []byte, key uint64) (int, error) {
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

func (u uint64Key) EncodeNonTerminal(buffer []byte, key uint64) (int, error) {
	return u.Encode(buffer, key)
}

func (u uint64Key) DecodeNonTerminal(buffer []byte) (int, uint64, error) {
	return u.Decode(buffer)
}

func (u uint64Key) SizeNonTerminal(key uint64) int {
	return u.Size(key)
}

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

func (stringKey) EncodeNonTerminal(buffer []byte, key string) (int, error) {
	for i := range key {
		c := key[i]
		if c == StringDelimiter {
			return 0, fmt.Errorf("%w: string is not allowed to have the string delimiter (%c) in non terminal encodings of strings", ErrEncoding, StringDelimiter)
		}
		buffer[i] = c
	}

	return len(key) + 1, nil
}

func (stringKey) DecodeNonTerminal(buffer []byte) (int, string, error) {
	i := bytes.IndexByte(buffer, StringDelimiter)
	if i == -1 {
		return 0, "", fmt.Errorf("%w: not a valid non terminal buffer, no instances of the string delimiter %c found", ErrEncoding, StringDelimiter)
	}
	return i + 1, string(buffer[:i]), nil
}

func (stringKey) SizeNonTerminal(key string) int { return len(key) + 1 }

func (stringKey) Stringify(key string) string {
	return key
}

func (stringKey) KeyType() string {
	return "string"
}
