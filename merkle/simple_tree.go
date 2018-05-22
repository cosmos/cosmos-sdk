/*
Computes a deterministic minimal height merkle tree hash.
If the number of items is not a power of two, some leaves
will be at different levels. Tries to keep both sides of
the tree the same size, but the left may be one greater.

Use this for short deterministic trees, such as the validator list.
For larger datasets, use IAVLTree.

                        *
                       / \
                     /     \
                   /         \
                 /             \
                *               *
               / \             / \
              /   \           /   \
             /     \         /     \
            *       *       *       h6
           / \     / \     / \
          h0  h1  h2  h3  h4  h5

*/

package merkle

import (
	"golang.org/x/crypto/ripemd160"
)

func SimpleHashFromTwoHashes(left []byte, right []byte) []byte {
	var hasher = ripemd160.New()
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

func SimpleHashFromHashes(hashes [][]byte) []byte {
	// Recursive impl.
	switch len(hashes) {
	case 0:
		return nil
	case 1:
		return hashes[0]
	default:
		left := SimpleHashFromHashes(hashes[:(len(hashes)+1)/2])
		right := SimpleHashFromHashes(hashes[(len(hashes)+1)/2:])
		return SimpleHashFromTwoHashes(left, right)
	}
}

// NOTE: Do not implement this, use SimpleHashFromByteslices instead.
// type Byteser interface { Bytes() []byte }
// func SimpleHashFromBytesers(items []Byteser) []byte { ... }

func SimpleHashFromByteslices(bzs [][]byte) []byte {
	hashes := make([][]byte, len(bzs))
	for i, bz := range bzs {
		hashes[i] = SimpleHashFromBytes(bz)
	}
	return SimpleHashFromHashes(hashes)
}

func SimpleHashFromBytes(bz []byte) []byte {
	hasher := ripemd160.New()
	hasher.Write(bz)
	return hasher.Sum(nil)
}

func SimpleHashFromHashers(items []Hasher) []byte {
	hashes := make([][]byte, len(items))
	for i, item := range items {
		hash := item.Hash()
		hashes[i] = hash
	}
	return SimpleHashFromHashes(hashes)
}

func SimpleHashFromMap(m map[string]Hasher) []byte {
	sm := NewSimpleMap()
	for k, v := range m {
		sm.Set(k, v)
	}
	return sm.Hash()
}
