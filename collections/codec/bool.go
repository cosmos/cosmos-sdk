package codec

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// NewBoolKey returns a new boolKey codec that can handle encoding and decoding boolean values.
// This codec can work with any type `T` that is derived from a boolean (~bool).
func NewBoolKey[T ~bool]() NameableKeyCodec[T] {
	return boolKey[T]{}
}

// boolKey implements the KeyCodec interface for boolean values.
// It provides methods for encoding and decoding boolean values as keys.
type boolKey[T ~bool] struct{}

// Encode encodes the boolean value into the first byte of the buffer.
// If the value is true, the buffer gets the byte value `0x1`, otherwise `0x0`.
// It returns the number of bytes written (always 1) or an error if something goes wrong.
func (b boolKey[T]) Encode(buffer []byte, key T) (int, error) {
	if key {
		buffer[0] = 0x1 // Encode true as 0x1.
		return 1, nil
	}
	buffer[0] = 0x0 // Encode false as 0x0.
	return 1, nil
}

// Decode reads the first byte from the buffer and converts it into a boolean value.
// If the byte is `0x0`, it decodes to false; if `0x1`, it decodes to true.
// It returns the number of bytes read (always 1), the decoded value, and an error if decoding fails.
func (b boolKey[T]) Decode(buffer []byte) (int, T, error) {
	if len(buffer) == 0 {
		// Error if the buffer length is less than 1.
		return 0, false, fmt.Errorf("%w: wanted size to be at least 1", ErrEncoding)
	}
	switch buffer[0] {
	case 0:
		// If the first byte is 0, return false.
		return 1, false, nil
	case 1:
		// If the first byte is 1, return true.
		return 1, true, nil
	default:
		// Error for any other invalid byte.
		return 0, false, fmt.Errorf("%w: invalid bool value: %d", ErrEncoding, buffer[0])
	}
}

// Size returns the size of the buffer needed to store the boolean value, which is always 1 byte.
func (b boolKey[T]) Size(_ T) int {
	return 1
}

// EncodeJSON converts the boolean value into its JSON representation (either "true" or "false").
// It returns the encoded JSON bytes or an error if encoding fails.
func (b boolKey[T]) EncodeJSON(value T) ([]byte, error) {
	return json.Marshal(value)
}

// DecodeJSON decodes JSON-encoded data into a boolean value.
// It returns the decoded value or an error if decoding fails.
func (b boolKey[T]) DecodeJSON(buffer []byte) (T, error) {
	var t T
	// Use json.Unmarshal to decode the buffer into the boolean type.
	err := json.Unmarshal(buffer, &t)
	return t, err
}

// Stringify converts the boolean value into a string ("true" or "false").
// This is useful for debugging or logging purposes.
func (b boolKey[T]) Stringify(key T) string {
	return strconv.FormatBool((bool)(key))
}

// KeyType returns a string identifier for the type of the key, which is "bool" in this case.
func (b boolKey[T]) KeyType() string {
	return "bool"
}

// EncodeNonTerminal is used in multipart keys. It behaves the same as the regular Encode method,
// writing the boolean value into the buffer.
func (b boolKey[T]) EncodeNonTerminal(buffer []byte, key T) (int, error) {
	return b.Encode(buffer, key)
}

// DecodeNonTerminal is used in multipart keys. It behaves the same as the regular Decode method,
// reading the boolean value from the buffer.
func (b boolKey[T]) DecodeNonTerminal(buffer []byte) (int, T, error) {
	return b.Decode(buffer)
}

// SizeNonTerminal returns the size of the buffer required to store the boolean value in multipart keys.
// It behaves the same as the regular Size method, returning 1 byte.
func (b boolKey[T]) SizeNonTerminal(key T) int {
	return b.Size(key)
}

// WithName wraps the current codec with a name. It returns a NamedKeyCodec, which contains
// the original codec along with a name for identification.
func (b boolKey[T]) WithName(name string) KeyCodec[T] {
	return NamedKeyCodec[T]{KeyCodec: b, Name: name}
}
