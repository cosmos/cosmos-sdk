package merkle

import (
	"github.com/tendermint/go-crypto/tmhash"
	cmn "github.com/tendermint/tmlibs/common"
)

// Merkle tree from a map.
// Leaves are `hash(key) | hash(value)`.
// Leaves are sorted before Merkle hashing.
type simpleMap struct {
	kvs    cmn.KVPairs
	sorted bool
}

func newSimpleMap() *simpleMap {
	return &simpleMap{
		kvs:    nil,
		sorted: false,
	}
}

// Set hashes the key and value and appends it to the kv pairs.
func (sm *simpleMap) Set(key string, value Hasher) {
	sm.sorted = false

	// The value is hashed, so you can
	// check for equality with a cached value (say)
	// and make a determination to fetch or not.
	vhash := value.Hash()

	sm.kvs = append(sm.kvs, cmn.KVPair{
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
func (sm *simpleMap) KVPairs() cmn.KVPairs {
	sm.Sort()
	kvs := make(cmn.KVPairs, len(sm.kvs))
	copy(kvs, sm.kvs)
	return kvs
}

//----------------------------------------

// A local extension to KVPair that can be hashed.
// Key and value are length prefixed and concatenated,
// then hashed.
type KVPair cmn.KVPair

func (kv KVPair) Hash() []byte {
	hasher := tmhash.New()
	err := encodeByteSlice(hasher, kv.Key)
	if err != nil {
		panic(err)
	}
	err = encodeByteSlice(hasher, kv.Value)
	if err != nil {
		panic(err)
	}
	return hasher.Sum(nil)
}

func hashKVPairs(kvs cmn.KVPairs) []byte {
	kvsH := make([]Hasher, len(kvs))
	for i, kvp := range kvs {
		kvsH[i] = KVPair(kvp)
	}
	return SimpleHashFromHashers(kvsH)
}
