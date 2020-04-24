package rootmulti

import (
	"bytes"

	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/libs/kv"
)

// Merkle tree from a map.
// Leaves are `hash(key) | hash(value)`.
// Leaves are sorted before Merkle hashing.
type simpleMap struct {
	kvs    kv.Pairs
	sorted bool
}

func newSimpleMap() *simpleMap {
	return &simpleMap{
		kvs:    nil,
		sorted: false,
	}
}

// Set creates a kv pair of the key and the hash of the value,
// and then appends it to simpleMap's kv pairs.
func (sm *simpleMap) Set(key string, value []byte) {
	sm.sorted = false

	// The value is hashed, so you can
	// check for equality with a cached value (say)
	// and make a determination to fetch or not.
	vhash := tmhash.Sum(value)

	sm.kvs = append(sm.kvs, kv.Pair{
		Key:   []byte(key),
		Value: vhash,
	})
}

// Hash Merkle root hash of items sorted by key
// (UNSTABLE: and by value too if duplicate key).
func (sm *simpleMap) Hash() []byte {
	sm.Sort()
	return hashKVPairs(sm.kvs)
}

func (sm *simpleMap) Sort() {
	if sm.sorted {
		return
	}
	sm.kvs.Sort()
	sm.sorted = true
}

// Returns a copy of sorted KVPairs.
// NOTE these contain the hashed key and value.
func (sm *simpleMap) KVPairs() kv.Pairs {
	sm.Sort()
	kvs := make(kv.Pairs, len(sm.kvs))
	copy(kvs, sm.kvs)
	return kvs
}

//----------------------------------------

// A local extension to KVPair that can be hashed.
// Key and value are length prefixed and concatenated,
// then hashed.
type KVPair kv.Pair

// Bytes returns key || value, with both the
// key and value length prefixed.
func (kv KVPair) Bytes() []byte {
	var b bytes.Buffer
	err := amino.EncodeByteSlice(&b, kv.Key)
	if err != nil {
		panic(err)
	}
	err = amino.EncodeByteSlice(&b, kv.Value)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func hashKVPairs(kvs kv.Pairs) []byte {
	kvsH := make([][]byte, len(kvs))
	for i, kvp := range kvs {
		kvsH[i] = KVPair(kvp).Bytes()
	}
	return merkle.SimpleHashFromByteSlices(kvsH)
}
