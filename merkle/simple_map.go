package merkle

import (
	cmn "github.com/tendermint/tmlibs/common"
	"golang.org/x/crypto/ripemd160"
)

type SimpleMap struct {
	kvs    cmn.KVPairs
	sorted bool
}

func NewSimpleMap() *SimpleMap {
	return &SimpleMap{
		kvs:    nil,
		sorted: false,
	}
}

func (sm *SimpleMap) Set(key string, value Hasher) {
	sm.sorted = false

	// Hash the key to blind it... why not?
	khash := SimpleHashFromBytes([]byte(key))

	// And the value is hashed too, so you can
	// check for equality with a cached value (say)
	// and make a determination to fetch or not.
	vhash := value.Hash()

	sm.kvs = append(sm.kvs, cmn.KVPair{
		Key:   khash,
		Value: vhash,
	})
}

// Merkle root hash of items sorted by key
// (UNSTABLE: and by value too if duplicate key).
func (sm *SimpleMap) Hash() []byte {
	sm.Sort()
	return hashKVPairs(sm.kvs)
}

func (sm *SimpleMap) Sort() {
	if sm.sorted {
		return
	}
	sm.kvs.Sort()
	sm.sorted = true
}

// Returns a copy of sorted KVPairs.
func (sm *SimpleMap) KVPairs() cmn.KVPairs {
	sm.Sort()
	kvs := make(cmn.KVPairs, len(sm.kvs))
	copy(kvs, sm.kvs)
	return kvs
}

//----------------------------------------

// A local extension to KVPair that can be hashed.
type KVPair cmn.KVPair

func (kv KVPair) Hash() []byte {
	hasher := ripemd160.New()
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
	kvsH := make([]Hasher, 0, len(kvs))
	for _, kvp := range kvs {
		kvsH = append(kvsH, KVPair(kvp))
	}
	return SimpleHashFromHashers(kvsH)
}
