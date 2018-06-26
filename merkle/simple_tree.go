package merkle

import (
	"github.com/tendermint/go-crypto/tmhash"
)

// SimpleHashFromTwoHashes is the basic operation of the Merkle tree: Hash(left | right).
func SimpleHashFromTwoHashes(left, right []byte) []byte {
	var hasher = tmhash.New()
	err := encodeByteSlice(hasher, left)
	if err != nil {
			panic(err)
		}
	err = encodeByteSlice(hasher, right)
	if err != nil {
			panic(err)
		}
	return hasher.Sum(nil)
}

// SimpleHashFromHashers computes a Merkle tree from items that can be hashed.
func SimpleHashFromHashers(items []Hasher) []byte {
	hashes := make([][]byte, len(items))
	for i, item := range items {
		hash := item.Hash()
		hashes[i] = hash
	}
	return simpleHashFromHashes(hashes)
}

// SimpleHashFromMap computes a Merkle tree from sorted map.
// Like calling SimpleHashFromHashers with
// `item = []byte(Hash(key) | Hash(value))`,
// sorted by `item`.
func SimpleHashFromMap(m map[string]Hasher) []byte {
	sm := newSimpleMap()
	for k, v := range m {
		sm.Set(k, v)
	}
	return sm.Hash()
}

//----------------------------------------------------------------

// Expects hashes!
func simpleHashFromHashes(hashes [][]byte) []byte {
	// Recursive impl.
	switch len(hashes) {
	case 0:
		return nil
	case 1:
		return hashes[0]
	default:
		left := simpleHashFromHashes(hashes[:(len(hashes)+1)/2])
		right := simpleHashFromHashes(hashes[(len(hashes)+1)/2:])
		return SimpleHashFromTwoHashes(left, right)
	}
}
