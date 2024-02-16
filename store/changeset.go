package store

import (
	"bytes"

	"cosmossdk.io/store/v2/internal/encoding"
)

// KVPair defines a key-value pair with additional metadata that is used to
// track writes. Deletion can be denoted by a nil value or explicitly by the
// Delete field.
type KVPair struct {
	Key      []byte
	Value    []byte
	StoreKey string // Optional for snapshot restore
}

type KVPairs []KVPair

// Changeset defines a set of KVPair entries by maintaining a map from store key
// to a slice of KVPair objects.
type Changeset struct {
	Pairs map[string]KVPairs
}

func NewChangeset() *Changeset {
	return &Changeset{
		Pairs: make(map[string]KVPairs),
	}
}

func NewChangesetWithPairs(pairs map[string]KVPairs) *Changeset {
	return &Changeset{
		Pairs: pairs,
	}
}

// Size returns the number of key-value pairs in the batch.
func (cs *Changeset) Size() int {
	cnt := 0
	for _, pairs := range cs.Pairs {
		cnt += len(pairs)
	}

	return cnt
}

// Add adds a key-value pair to the ChangeSet.
func (cs *Changeset) Add(storeKey string, key, value []byte) {
	cs.Pairs[storeKey] = append(cs.Pairs[storeKey], KVPair{
		Key:      key,
		Value:    value,
		StoreKey: storeKey,
	})
}

// AddKVPair adds a KVPair to the ChangeSet.
func (cs *Changeset) AddKVPair(storeKey string, pair KVPair) {
	cs.Pairs[storeKey] = append(cs.Pairs[storeKey], pair)
}

// Merge merges the provided Changeset argument into the receiver. This may be
// useful when you have a Changeset that only pertains to a single store key,
// i.e. a map of size one, and you want to merge it into another.
func (cs *Changeset) Merge(other *Changeset) {
	for storeKey, pairs := range other.Pairs {
		cs.Pairs[storeKey] = append(cs.Pairs[storeKey], pairs...)
	}
}

// encodedSize returns the size of the encoded Changeset.
func (cs *Changeset) encodedSize() int {
	size := encoding.EncodeUvarintSize(uint64(len(cs.Pairs)))
	for storeKey, pairs := range cs.Pairs {
		size += encoding.EncodeBytesSize([]byte(storeKey))
		size += encoding.EncodeUvarintSize(uint64(len(pairs)))
		for _, pair := range pairs {
			size += encoding.EncodeBytesSize(pair.Key)
			size += encoding.EncodeBytesSize(pair.Value)
		}
	}
	return size
}

// Marshal returns the encoded byte representation of Changeset.
// NOTE: The Changeset is encoded as follows:
// - number of store keys (uvarint)
// - for each store key:
// -- store key (bytes)
// -- number of pairs (uvarint)
// -- for each pair:
// --- key (bytes)
// --- value (bytes)
func (cs *Changeset) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(cs.encodedSize())

	if err := encoding.EncodeUvarint(&buf, uint64(len(cs.Pairs))); err != nil {
		return nil, err
	}
	for storeKey, pairs := range cs.Pairs {
		if err := encoding.EncodeBytes(&buf, []byte(storeKey)); err != nil {
			return nil, err
		}
		if err := encoding.EncodeUvarint(&buf, uint64(len(pairs))); err != nil {
			return nil, err
		}
		for _, pair := range pairs {
			if err := encoding.EncodeBytes(&buf, pair.Key); err != nil {
				return nil, err
			}
			if err := encoding.EncodeBytes(&buf, pair.Value); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

// Unmarshal decodes the Changeset from the given byte slice.
func (cs *Changeset) Unmarshal(buf []byte) error {
	storeCount, n, err := encoding.DecodeUvarint(buf)
	if err != nil {
		return err
	}
	buf = buf[n:]

	cs.Pairs = make(map[string]KVPairs, storeCount)
	for i := uint64(0); i < storeCount; i++ {
		storeKey, n, err := encoding.DecodeBytes(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]

		pairCount, n, err := encoding.DecodeUvarint(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]

		pairs := make(KVPairs, pairCount)
		for j := uint64(0); j < pairCount; j++ {
			key, n, err := encoding.DecodeBytes(buf)
			if err != nil {
				return err
			}
			buf = buf[n:]

			value, n, err := encoding.DecodeBytes(buf)
			if err != nil {
				return err
			}
			buf = buf[n:]

			pairs[j] = KVPair{
				Key:      key,
				Value:    value,
				StoreKey: string(storeKey),
			}
		}
		cs.Pairs[string(storeKey)] = pairs
	}

	return nil
}
