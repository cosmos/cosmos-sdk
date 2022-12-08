package types

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

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

// assertNoCommonPrefix will panic if there are two keys: k1 and k2 in keys, such that
// k1 is a prefix of k2
func assertNoCommonPrefix(keys []string) {
	sorted := make([]string, len(keys))
	copy(sorted, keys)
	sort.Strings(sorted)
	for i := 1; i < len(sorted); i++ {
		if strings.HasPrefix(sorted[i], sorted[i-1]) {
			panic(fmt.Sprint("Potential key collision between KVStores:", sorted[i], " - ", sorted[i-1]))
		}
	}
}

// Codec defines a interface needed for the store package to marshal data
type Codec interface {
	// Marshal returns binary encoding of v.
	Marshal(proto.Message) ([]byte, error)

	// MarshalLengthPrefixed returns binary encoding of v with bytes length prefix.
	MarshalLengthPrefixed(proto.Message) ([]byte, error)

	// Unmarshal parses the data encoded with Marshal method and stores the result
	// in the value pointed to by v.
	Unmarshal(bz []byte, ptr proto.Message) error

	// Unmarshal parses the data encoded with UnmarshalLengthPrefixed method and stores
	// the result in the value pointed to by v.
	UnmarshalLengthPrefixed(bz []byte, ptr proto.Message) error
}

// ============= TestCodec =============
// TestCodec defines a codec that utilizes Protobuf for both binary and JSON
// encoding.
type TestCodec struct{}

var _ Codec = &TestCodec{}

func NewTestCodec() Codec {
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

// Unmarshal implements BinaryMarshaler.Unmarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterface
func (pc *TestCodec) Unmarshal(bz []byte, ptr proto.Message) error {
	err := proto.Unmarshal(bz, ptr)
	if err != nil {
		return err
	}

	return nil
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
