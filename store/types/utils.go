package types

import (
	"bytes"
	"encoding/binary"
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/types/kv"
	proto "github.com/cosmos/gogoproto/proto"
)

// KVStorePrefixIterator iterates over all the keys with a certain prefix in ascending order
func KVStorePrefixIterator(kvs KVStore, prefix []byte) Iterator {
	return kvs.Iterator(prefix, PrefixEndBytes(prefix))
}

// KVStoreReversePrefixIterator iterates over all the keys with a certain prefix in descending order.
func KVStoreReversePrefixIterator(kvs KVStore, prefix []byte) Iterator {
	return kvs.ReverseIterator(prefix, PrefixEndBytes(prefix))
}

// DiffKVStores compares two KVstores and returns all the key/value pairs
// that differ from one another. It also skips value comparison for a set of provided prefixes.
func DiffKVStores(a KVStore, b KVStore, prefixesToSkip [][]byte) (kvAs, kvBs []kv.Pair) {
	iterA := a.Iterator(nil, nil)

	defer iterA.Close()

	iterB := b.Iterator(nil, nil)

	defer iterB.Close()

	for {
		if !iterA.Valid() && !iterB.Valid() {
			return kvAs, kvBs
		}

		var kvA, kvB kv.Pair
		if iterA.Valid() {
			kvA = kv.Pair{Key: iterA.Key(), Value: iterA.Value()}

			iterA.Next()
		}

		if iterB.Valid() {
			kvB = kv.Pair{Key: iterB.Key(), Value: iterB.Value()}
		}

		compareValue := true

		for _, prefix := range prefixesToSkip {
			// Skip value comparison if we matched a prefix
			if bytes.HasPrefix(kvA.Key, prefix) {
				compareValue = false
				break
			}
		}

		if !compareValue {
			// We're skipping this key due to an exclusion prefix.  If it's present in B, iterate past it.  If it's
			// absent don't iterate.
			if bytes.Equal(kvA.Key, kvB.Key) {
				iterB.Next()
			}
			continue
		}

		// always iterate B when comparing
		iterB.Next()

		if !bytes.Equal(kvA.Key, kvB.Key) || !bytes.Equal(kvA.Value, kvB.Value) {
			kvAs = append(kvAs, kvA)
			kvBs = append(kvBs, kvB)
		}
	}
}

// PrefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func PrefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		}

		end = end[:len(end)-1]

		if len(end) == 0 {
			end = nil
			break
		}
	}

	return end
}

// InclusiveEndBytes returns the []byte that would end a
// range query such that the input would be included
func InclusiveEndBytes(inclusiveBytes []byte) []byte {
	return append(inclusiveBytes, byte(0x00))
}

// Marshaler defines a interface needed for the store package to marshal data
type Marshaler interface {
	// Marshal returns binary encoding of v.
	Marshal(proto.Message) ([]byte, error)

	// MarshalLengthPrefixed returns binary encoding of v with bytes length prefix.
	MarshalLengthPrefixed(proto.Message) ([]byte, error)

	// Unmarshal parses the data encoded with UnmarshalLengthPrefixed method and stores
	// the result in the value pointed to by v.
	UnmarshalLengthPrefixed(bz []byte, ptr proto.Message) error
}

// ============= TestCodec =============
// TestCodec defines a codec that utilizes Protobuf for both binary and JSON
// encoding.
type TestCodec struct{}

var _ Marshaler = &TestCodec{}

func NewTestCodec() Marshaler {
	return &TestCodec{}
}

// Marshal implements BinaryMarshaler.Marshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterface
func (pc *TestCodec) Marshal(o proto.Message) ([]byte, error) {
	// Size() check can catch the typed nil value.
	if o == nil || proto.Size(o) == 0 {
		// return empty bytes instead of nil, because nil has special meaning in places like store.Set
		return []byte{}, nil
	}
	return proto.Marshal(o)
}

// MarshalLengthPrefixed implements BinaryMarshaler.MarshalLengthPrefixed method.
func (pc *TestCodec) MarshalLengthPrefixed(o proto.Message) ([]byte, error) {
	bz, err := pc.Marshal(o)
	if err != nil {
		return nil, err
	}

	var sizeBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(sizeBuf[:], uint64(proto.Size(o)))
	return append(sizeBuf[:n], bz...), nil
}

// UnmarshalLengthPrefixed implements BinaryMarshaler.UnmarshalLengthPrefixed method.
func (pc *TestCodec) UnmarshalLengthPrefixed(bz []byte, ptr proto.Message) error {
	size, n := binary.Uvarint(bz)
	if n < 0 {
		return fmt.Errorf("invalid number of bytes read from length-prefixed encoding: %d", n)
	}

	if size > uint64(len(bz)-n) {
		return fmt.Errorf("not enough bytes to read; want: %v, got: %v", size, len(bz)-n)
	} else if size < uint64(len(bz)-n) {
		return fmt.Errorf("too many bytes to read; want: %v, got: %v", size, len(bz)-n)
	}

	bz = bz[n:]
	return proto.Unmarshal(bz, ptr)
}
