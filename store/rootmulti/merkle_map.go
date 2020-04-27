package rootmulti

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/libs/kv"
)

// MerkleMap is a merkle tree from a map.
// Leaves are `hash(key) | hash(value)`.
// Leaves are sorted before Merkle hashing.
type merkleMap struct {
	kvs    kv.Pairs
	sorted bool
}

func newMerkleMap() *merkleMap {
	return &merkleMap{
		kvs:    nil,
		sorted: false,
	}
}

// Set creates a kv pair of the key and the hash of the value,
// and then appends it to merkleMap's kv pairs.
func (sm *merkleMap) Set(key string, value []byte) {
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
func (sm *merkleMap) Hash() []byte {
	sm.Sort()
	return hashKVPairs(sm.kvs)
}

func (sm *merkleMap) Sort() {
	if sm.sorted {
		return
	}
	sm.kvs.Sort()
	sm.sorted = true
}

// Returns a copy of sorted KVPairs.
// NOTE these contain the hashed key and value.
func (sm *merkleMap) KVPairs() kv.Pairs {
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
	err := encodeByteSlice(&b, kv.Key)
	if err != nil {
		panic(err)
	}
	err = encodeByteSlice(&b, kv.Value)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

// EncodeByteSlice encodes a byte slice with its length prefixed
func encodeByteSlice(w io.Writer, bz []byte) (err error) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], uint64(len(bz)))
	_, err = w.Write(buf[0:n])
	if err != nil {
		return
	}
	_, err = w.Write(bz)
	return
}

// hashKVPairs hashes a KVPair and creates a merkle tree where the leaves are byte slices
func hashKVPairs(kvs kv.Pairs) []byte {
	kvsH := make([][]byte, len(kvs))
	for i, kvp := range kvs {
		kvsH[i] = KVPair(kvp).Bytes()
	}
	return merkle.SimpleHashFromByteSlices(kvsH)
}
