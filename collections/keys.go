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
	// Uint32Key can be used to encode uint32 keys. Encoding is big endian to retain ordering.
	Uint32Key KeyCodec[uint32] = uint32Key[uint32]{}
	// Uint16Key can be used to encode uint16 keys. Encoding is big endian to retain ordering.
	Uint16Key KeyCodec[uint16] = uint16Key[uint16]{}
	// BoolKey can be used to encode booleans. It uses a single byte to represent the boolean.
	// 0x0 is used to represent false, and 0x1 is used to represent true.
	BoolKey KeyCodec[bool] = boolKey[bool]{}
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

func (uint64Key) EncodeJSON(value uint64) ([]byte, error) { return uint64EncodeJSON(value) }

func (uint64Key) DecodeJSON(b []byte) (uint64, error) { return uint64DecodeJSON(b, 64) }

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

type uint32Key[T ~uint32] struct{}

func (uint32Key[T]) Encode(buffer []byte, key T) (int, error) {
	binary.BigEndian.PutUint32(buffer, (uint32)(key))
	return 4, nil
}

func (uint32Key[T]) Decode(buffer []byte) (int, T, error) {
	if len(buffer) < 4 {
		return 0, 0, fmt.Errorf("%w: expected buffer of size 4", ErrEncoding)
	}
	return 4, (T)(binary.BigEndian.Uint32(buffer)), nil
}

func (uint32Key[T]) Size(_ T) int { return 4 }

func (uint32Key[T]) EncodeJSON(value T) ([]byte, error) { return uint64EncodeJSON((uint64)(value)) }

func (uint32Key[T]) DecodeJSON(b []byte) (T, error) {
	u, err := uint64DecodeJSON(b, 32)
	if err != nil {
		return 0, err
	}
	return (T)(u), nil
}

func (uint32Key[T]) Stringify(key T) string { return strconv.FormatUint(uint64(key), 10) }

func (uint32Key[T]) KeyType() string { return "uint32" }

func (u uint32Key[T]) EncodeNonTerminal(buffer []byte, key T) (int, error) {
	return u.Encode(buffer, key)
}

func (u uint32Key[T]) DecodeNonTerminal(buffer []byte) (int, T, error) { return u.Decode(buffer) }

func (uint32Key[T]) SizeNonTerminal(_ T) int { return 4 }

type uint16Key[T ~uint16] struct{}

func (uint16Key[T]) Encode(buffer []byte, key T) (int, error) {
	binary.BigEndian.PutUint16(buffer, (uint16)(key))
	return 2, nil
}

func (uint16Key[T]) Decode(buffer []byte) (int, T, error) {
	if len(buffer) < 2 {
		return 0, 0, fmt.Errorf("%w: invalid buffer size, wanted at least 2", ErrEncoding)
	}
	return 2, (T)(binary.BigEndian.Uint16(buffer)), nil
}

func (uint16Key[T]) Size(key T) int { return 2 }

func (uint16Key[T]) EncodeJSON(value T) ([]byte, error) { return uint64EncodeJSON((uint64)(value)) }

func (uint16Key[T]) DecodeJSON(b []byte) (T, error) {
	u, err := uint64DecodeJSON(b, 16)
	if err != nil {
		return 0, err
	}
	return (T)(u), nil
}

func (uint16Key[T]) Stringify(key T) string { return strconv.FormatUint((uint64)(key), 10) }

func (uint16Key[T]) KeyType() string { return "uint16" }

func (u uint16Key[T]) EncodeNonTerminal(buffer []byte, key T) (int, error) {
	return u.Encode(buffer, key)
}

func (u uint16Key[T]) DecodeNonTerminal(buffer []byte) (int, T, error) { return u.Decode(buffer) }

func (u uint16Key[T]) SizeNonTerminal(key T) int { return u.Size(key) }

type boolKey[T ~bool] struct{}

func (b boolKey[T]) Encode(buffer []byte, key T) (int, error) {
	if key {
		buffer[0] = 0x1
		return 1, nil
	} else {
		buffer[0] = 0x0
		return 1, nil
	}
}

func (b boolKey[T]) Decode(buffer []byte) (int, T, error) {
	if len(buffer) == 0 {
		return 0, false, fmt.Errorf("%w: wanted size to be at least 1", ErrEncoding)
	}
	switch buffer[0] {
	case 0:
		return 1, false, nil
	case 1:
		return 1, true, nil
	default:
		return 0, false, fmt.Errorf("%w: invalid bool value: %d", ErrEncoding, buffer[0])
	}
}

func (b boolKey[T]) Size(key T) int { return 1 }

func (b boolKey[T]) EncodeJSON(value T) ([]byte, error) {
	return []byte(`"` + strconv.FormatBool((bool)(value)) + `"`), nil
}

func (b boolKey[T]) DecodeJSON(buffer []byte) (T, error) {
	s, err := StringKey.DecodeJSON(buffer)
	if err != nil {
		return false, err
	}
	k, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return (T)(k), nil
}

func (b boolKey[T]) Stringify(key T) string {
	return strconv.FormatBool((bool)(key))
}

func (b boolKey[T]) KeyType() string {
	return "bool"
}

func (b boolKey[T]) EncodeNonTerminal(buffer []byte, key T) (int, error) {
	return b.Encode(buffer, key)
}

func (b boolKey[T]) DecodeNonTerminal(buffer []byte) (int, T, error) {
	return b.Decode(buffer)
}

func (b boolKey[T]) SizeNonTerminal(key T) int {
	return b.Size(key)
}
