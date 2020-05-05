package rootmulti

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/libs/kv"
)

// merkleMap defines a merkle-ized tree from a map. Leave values are treated as
// hash(key) | hash(value). Leaves are sorted before Merkle hashing.
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

// set creates a kv.Pair from the provided key and value. The value is hashed prior
// to creating a kv.Pair. The created kv.Pair is appended to the merkleMap's slice
// of kv.Pairs. Whenever called, the merkleMap must be resorted.
func (sm *merkleMap) set(key string, value []byte) {
	sm.sorted = false

	// The value is hashed, so you can check for equality with a cached value (say)
	// and make a determination to fetch or not.
	vhash := tmhash.Sum(value)

	sm.kvs = append(sm.kvs, kv.Pair{
		Key:   []byte(key),
		Value: vhash,
	})
}

// hash returns the merkle root of items sorted by key. Note, it is unstable.
func (sm *merkleMap) hash() []byte {
	sm.sort()
	return hashKVPairs(sm.kvs)
}

func (sm *merkleMap) sort() {
	if sm.sorted {
		return
	}

	sm.kvs.Sort()
	sm.sorted = true
}

// kvPair defines a type alias for kv.Pair so that we can create bytes to hash
// when constructing the merkle root. Note, key and values are both length-prefixed.
type kvPair kv.Pair

// bytes returns a byte slice representation of the kvPair where the key and value
// are length-prefixed.
func (kv kvPair) bytes() []byte {
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

func encodeByteSlice(w io.Writer, bz []byte) error {
	var buf [8]byte
	n := binary.PutUvarint(buf[:], uint64(len(bz)))

	_, err := w.Write(buf[:n])
	if err != nil {
		return err
	}

	_, err = w.Write(bz)
	return err
}

// hashKVPairs hashes a kvPair and creates a merkle tree where the leaves are
// byte slices.
func hashKVPairs(kvs kv.Pairs) []byte {
	kvsH := make([][]byte, len(kvs))
	for i, kvp := range kvs {
		kvsH[i] = kvPair(kvp).bytes()
	}

	return merkle.SimpleHashFromByteSlices(kvsH)
}
