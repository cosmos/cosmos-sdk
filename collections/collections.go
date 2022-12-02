package collections

import (
	"bytes"
	"fmt"
	"math"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// StorageProvider represents sdk.Context
// it is used to avoid to reduce dependencies.
type StorageProvider interface {
	// KVStore returns a KVStore given its StoreKey.
	KVStore(key storetypes.StoreKey) storetypes.KVStore
}

// ErrNotFound defines an error returned when
// a key was not found.
type ErrNotFound struct {
	// HumanizedKey represents the human-readable key.
	HumanizedKey string
	// RawKey represents the raw key in byte format.
	RawKey []byte
	// ValueType represents the type which was not found.
	ValueType string
}

// Error implements the error interface.
func (e ErrNotFound) Error() string {
	return fmt.Sprintf("collections: key not found for %s '%s'", e.ValueType, e.HumanizedKey)
}

// Is implements error.Is interface
func (e ErrNotFound) Is(err error) bool {
	self, ok := err.(ErrNotFound)
	if !ok {
		return false
	}
	return bytes.Equal(self.RawKey, e.RawKey) &&
		self.HumanizedKey == e.HumanizedKey &&
		self.ValueType == e.ValueType
}

// Prefix defines a segregation namespace
// for specific collections objects.
type Prefix struct {
	raw []byte // TODO(testinginprod): maybe add a humanized prefix?
}

// Bytes returns the raw Prefix bytes.
func (n Prefix) Bytes() []byte { return n.raw }

// NewPrefix returns a Prefix given the provided namespace identifier.
// Prefixes of the same module must not start
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
		prefix = c // maybe copy?
	}
	return Prefix{raw: prefix}
}

// KeyEncoder defines a generic interface which is implemented
// by types that are capable of encoding and decoding collections keys.
type KeyEncoder[T any] interface {
	// Encode encodes the type T into bytes.
	Encode(key T) ([]byte, error)
	// Decode decodes the given bytes back into T.
	// And it also must return the bytes of the buffer which were read.
	Decode(b []byte) (int, T, error)
	// Stringify returns a string representation of T.
	Stringify(key T) string
	// KeyType returns an identifier for the key.
	KeyType() string
}

// ValueEncoder defines a generic interface which is implemented
// by types that are capable of encoding and decoding collection values.
type ValueEncoder[T any] interface {
	// Encode encodes the value T into bytes.
	Encode(value T) ([]byte, error)
	// Decode returns the type T given its bytes representation.
	Decode(b []byte) (T, error)
	// Stringify returns a string representation of T.
	Stringify(value T) string
	// ValueType returns the identifier for the type.
	ValueType() string
}
