package collections

import (
	"errors"
	"math"
)

var (
	// ErrNotFound is returned when the provided key is not present in the StorageProvider.
	ErrNotFound = errors.New("collections: not found")
	// ErrEncoding is returned when something fails during key or value encoding/decoding.
	ErrEncoding = errors.New("collections: encoding error")
	// ErrConflict is returned when there are conflicts, for example in UniqueIndex.
	ErrConflict = errors.New("collections: conflict")
)

// collection is the interface that all collections support. It will eventually
// include methods for importing/exporting genesis data and schema
// reflection for clients.
type collection interface {
	// getName is the unique name of the collection within a schema. It must
	// match format specified by NameRegex.
	getName() string

	// getPrefix is the unique prefix of the collection within a schema.
	getPrefix() []byte

	genesisHandler
}

// Prefix defines a segregation namespace
// for specific collections objects.
type Prefix struct {
	raw []byte // TODO(testinginprod): maybe add a humanized prefix field?
}

// Bytes returns the raw Prefix bytes.
func (n Prefix) Bytes() []byte { return n.raw }

// NewPrefix returns a Prefix given the provided namespace identifier.
// In the same module, no prefixes should share the same starting bytes
// meaning that having two namespaces whose bytes representation is:
// p1 := []byte("prefix")
// p2 := []byte("prefix1")
// yields to iterations of p1 overlapping over p2.
// If a numeric prefix is provided, it must be between 0 and 255 (uint8).
// If out of bounds this function will panic.
// Reason for which this function is constrained to `int` instead of `uint8` is for
// API ergonomics, golang's type inference will infer int properly but not uint8
// meaning that developers would need to write NewPrefix(uint8(number)) for numeric
// prefixes.
func NewPrefix[T interface{ int | string | []byte }](identifier T) Prefix {
	i := any(identifier)
	var prefix []byte
	switch c := i.(type) {
	case int:
		if c > math.MaxUint8 || c < 0 {
			panic("invalid integer prefix value: must be between 0 and 255")
		}
		prefix = []byte{uint8(c)}
	case string:
		prefix = []byte(c)
	case []byte:
		identifierCopy := make([]byte, len(c))
		copy(identifierCopy, c)
		prefix = identifierCopy
	}
	return Prefix{raw: prefix}
}

// KeyCodec defines a generic interface which is implemented
// by types that are capable of encoding and decoding collections keys.
type KeyCodec[T any] interface {
	// Encode writes the key bytes into the buffer. Returns the number of
	// bytes written. The implementer must expect the buffer to be at least
	// of length equal to Size(K) for all encodings.
	// It must also return the number of written bytes which must be
	// equal to Size(K) for all encodings not involving varints.
	// In case of encodings involving varints then the returned
	// number of written bytes is allowed to be smaller than Size(K).
	Encode(buffer []byte, key T) (int, error)
	// Decode reads from the provided bytes buffer to decode
	// the key T. Returns the number of bytes read, the type T
	// or an error in case of decoding failure.
	Decode(buffer []byte) (int, T, error)
	// Size returns the buffer size need to encode key T in binary format.
	// The returned value must match what is computed by Encode for all
	// encodings except the ones involving varints. Varints are expected
	// to return the maximum varint bytes buffer length, at the risk of
	// over-estimating in order to pick the most performant path.
	Size(key T) int
	// EncodeJSON encodes the value as JSON.
	EncodeJSON(value T) ([]byte, error)
	// DecodeJSON decodes the provided JSON bytes into an instance of T.
	DecodeJSON(b []byte) (T, error)
	// Stringify returns a string representation of T.
	Stringify(key T) string
	// KeyType returns a string identifier for the type of the key.
	KeyType() string

	// MULTIPART keys

	// EncodeNonTerminal writes the key bytes into the buffer.
	// EncodeNonTerminal is used in multipart keys like Pair
	// when the part of the key being encoded is not the last one,
	// and there needs to be a way to distinguish after how many bytes
	// the first part of the key is finished. The buffer is expected to be
	// at least as big as SizeNonTerminal(key) returns. It returns
	// the amount of bytes written.
	EncodeNonTerminal(buffer []byte, key T) (int, error)
	// DecodeNonTerminal reads the buffer provided and returns
	// the key T. DecodeNonTerminal is used in multipart keys
	// like Pair when the part of the key being decoded is not the
	// last one. It returns the amount of bytes read.
	DecodeNonTerminal(buffer []byte) (int, T, error)
	// SizeNonTerminal returns the maximum size of the key K when used in
	// multipart keys like Pair.
	SizeNonTerminal(key T) int
}

// ValueCodec defines a generic interface which is implemented
// by types that are capable of encoding and decoding collection values.
type ValueCodec[T any] interface {
	// Encode encodes the value T into binary format.
	Encode(value T) ([]byte, error)
	// Decode returns the type T given its binary representation.
	Decode(b []byte) (T, error)
	// EncodeJSON encodes the value as JSON.
	EncodeJSON(value T) ([]byte, error)
	// DecodeJSON decodes the provided JSON bytes into an instance of T.
	DecodeJSON(b []byte) (T, error)
	// Stringify returns a string representation of T.
	Stringify(value T) string
	// ValueType returns the identifier for the type.
	ValueType() string
}
