package collections

import (
	"cosmossdk.io/collections/codec"
	"errors"
	"math"
)

var (
	// ErrNotFound is returned when the provided key is not present in the StorageProvider.
	ErrNotFound = errors.New("collections: not found")
	// ErrEncoding is returned when something fails during key or value encoding/decoding.
	ErrEncoding = codec.ErrEncoding
	// ErrConflict is returned when there are conflicts, for example in UniqueIndex.
	ErrConflict = errors.New("collections: conflict")
)

// KEYS

var (
	// Uint64Key can be used to encode uint64 keys. Encoding is big endian to retain ordering.
	Uint64Key = codec.NewUint64Key[uint64]()
	// Uint32Key can be used to encode uint32 keys. Encoding is big endian to retain ordering.
	Uint32Key = codec.NewUint32Key[uint32]()
	// Uint16Key can be used to encode uint16 keys. Encoding is big endian to retain ordering.
	Uint16Key = codec.NewUint16Key[uint16]()
	// Int64Key can be used to encode int64. Encoding retains ordering by toggling the MSB.
	Int64Key = codec.NewInt64Key[int64]()
	// StringKey can be used to encode string keys. The encoding just converts the string
	// to bytes.
	// Non-terminality in multipart keys is handled by appending the StringDelimiter,
	// this means that a string key when used as the non final part of a multipart key cannot
	// contain the StringDelimiter.
	// Lexicographical ordering is retained both in non and multipart keys.
	StringKey = codec.NewStringKeyCodec[string]()
	// BytesKey can be used to encode bytes keys. The encoding will just use
	// the provided bytes.
	// When used as the non-terminal part of a multipart key, we prefix the bytes key
	// with a single byte representing the length of the key. This means two things:
	// 1. When used in multipart keys the length can be at maximum 255 (max number that
	// can be represented with a single byte).
	// 2. When used in multipart keys the lexicographical ordering is lost due to the
	// length prefixing.
	// JSON encoding represents a bytes key as a hex encoded string.
	BytesKey = codec.NewBytesKey[[]byte]()
	// BoolKey can be used to encode booleans. It uses a single byte to represent the boolean.
	// 0x0 is used to represent false, and 0x1 is used to represent true.
	BoolKey = codec.NewBoolKey[bool]()
)

// VALUES

var (
	// Uint64Value implements a ValueCodec for uint64. It converts the uint64 to big endian bytes.
	// The JSON representation is the string format of uint64.
	Uint64Value = codec.KeyToValueCodec(Uint64Key)
	// StringValue implements a ValueCodec for string.
	StringValue = codec.KeyToValueCodec(StringKey)
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
