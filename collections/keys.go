package collections

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
)

var (
	// Uint64Key can be used to encode uint64 keys. Encoding is big endian to retain ordering.
	Uint64Key KeyCodec[uint64] = uint64Key{}
	// StringKey can be used to encode string keys. The encoding just converts the string
	// to bytes.
	// Non-terminality in multipart keys is handled by appending the StringDelimiter,
	// this means that a string key when used as the non final part of a multipart key cannot
	// contain the StringDelimiter.
	// Lexicographical ordering is retained both in non and multipart keys.
	StringKey KeyCodec[string] = stringKey{}
	// BytesKey can be used to encode bytes keys. The encoding will just use
	// the provided bytes.
	// When used as the non-terminal part of a multipart key, we prefix the bytes key
	// with a single byte representing the length of the key. This means two things:
	// 1. When used in multipart keys the length can be at maximum 255 (max number that
	// can be represented with a single byte).
	// 2. When used in multipart keys the lexicographical ordering is lost due to the
	// length prefixing.
	// JSON encoding represents a bytes key as a hex encoded string.
	BytesKey KeyCodec[[]byte] = bytesKey{}
)

const (
	// StringDelimiter defines the delimiter of a string key when used in non-terminal encodings.
	StringDelimiter uint8 = 0x0
	// MaxBytesKeyNonTerminalSize defines the maximum length of a bytes key encoded
	// using the BytesKey KeyCodec.
	MaxBytesKeyNonTerminalSize = math.MaxUint8
)

// errDecodeKeySize is a sentinel error.
var errDecodeKeySize = errors.New("decode error, wrong byte key size")

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

type bytesKey struct{}

func (b bytesKey) Encode(buffer []byte, key []byte) (int, error) {
	return copy(buffer, key), nil
}

func (bytesKey) Decode(buffer []byte) (int, []byte, error) {
	// todo: should we copy it? collections will just discard the buffer, so from coll POV is not needed.
	return len(buffer), buffer, nil
}

func (bytesKey) Size(key []byte) int {
	return len(key)
}

func (bytesKey) EncodeJSON(value []byte) ([]byte, error) {
	return StringKey.EncodeJSON(hex.EncodeToString(value))
}

func (bytesKey) DecodeJSON(b []byte) ([]byte, error) {
	hexBytes, err := StringKey.DecodeJSON(b)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(hexBytes)
}

func (b bytesKey) Stringify(key []byte) string {
	return fmt.Sprintf("hexBytes:%x", key)
}

func (b bytesKey) KeyType() string {
	return "bytes"
}

func (b bytesKey) EncodeNonTerminal(buffer []byte, key []byte) (int, error) {
	if len(key) > MaxBytesKeyNonTerminalSize {
		return 0, fmt.Errorf(
			"%w: bytes key non terminal size cannot exceed: %d, got: %d",
			ErrEncoding, MaxBytesKeyNonTerminalSize, len(key),
		)
	}

	buffer[0] = uint8(len(key))
	written := copy(buffer[1:], key)
	return written + 1, nil
}

func (bytesKey) DecodeNonTerminal(buffer []byte) (int, []byte, error) {
	l := len(buffer)
	if l == 0 {
		return 0, nil, fmt.Errorf("%w: bytes key non terminal decoding cannot have an empty buffer", ErrEncoding)
	}

	keyLength := int(buffer[0])
	if len(buffer[1:]) < keyLength {
		return 0, nil, fmt.Errorf(
			"%w: bytes key non terminal decoding isn't big enough, want at least: %d, got: %d",
			ErrEncoding, keyLength, len(buffer[1:]),
		)
	}
	return 1 + keyLength, buffer[1 : keyLength+1], nil
}

func (bytesKey) SizeNonTerminal(key []byte) int {
	return len(key) + 1
}
