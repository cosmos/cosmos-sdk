package collections

import (
	"bytes"
	"fmt"
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

// Namespace provide a storage byte prefix namespace
// which must be unique across storage objects belonging
// to a module.
// Besides this, once a namespace is defined: other module's
// namespaces cannot start with the same returned prefix to
// avoid namespace collisions, example:
// Namespace1=> "bank"
// Namespace2=> "bank2" <- namespace collision: Namespace1 starts with Namespace2
type Namespace struct {
	prefix []byte // TODO(testinginprod): maybe add a humanized prefix?
}

// Prefix returns the raw prefix bytes.
func (n Namespace) Prefix() []byte { return n.prefix }

// NewNamespace returns a Namespacer given the provided namespace identifier.
func NewNamespace[T interface{ uint8 | string | []byte }](identifier T) Namespace {
	i := any(identifier)
	var prefix []byte
	switch c := i.(type) {
	case uint8:
		prefix = []byte{c}
	case string:
		prefix = []byte(c)
	case []byte:
		prefix = c // maybe copy?
	}
	return Namespace{prefix: prefix}
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
