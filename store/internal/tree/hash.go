package tree

import (
	"crypto/sha256"
	"hash"
	"math/bits"
	"slices"
)

var (
	leafPrefix  = []byte{0}
	innerPrefix = []byte{1}
)

// HashFromByteSlices computes a Merkle tree where the leaves are the byte slices,
// in the provided order. It follows RFC-6962.
func HashFromByteSlices(items [][]byte) []byte {
	return hashFromByteSlices(sha256.New(), items)
}

// emptyHash is the sha256 hash of the empty string.
var emptyHash = [32]byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}

func hashFromByteSlices(sha hash.Hash, items [][]byte) []byte {
	switch len(items) {
	case 0:
		return slices.Clone(emptyHash[:])
	case 1:
		return leafHashOpt(sha, items[0])
	default:
		k := getSplitPoint(int64(len(items)))
		left := hashFromByteSlices(sha, items[:k])
		right := hashFromByteSlices(sha, items[k:])
		return innerHashOpt(sha, left, right)
	}
}

// leafHashOpt returns tmhash(0x00 || leaf)
func leafHashOpt(s hash.Hash, leaf []byte) []byte {
	s.Reset()
	s.Write(leafPrefix)
	s.Write(leaf)
	return s.Sum(nil)
}

func innerHashOpt(s hash.Hash, left, right []byte) []byte {
	s.Reset()
	s.Write(innerPrefix)
	s.Write(left)
	s.Write(right)
	return s.Sum(nil)
}

// getSplitPoint returns the largest power of 2 less than length
func getSplitPoint(length int64) int64 {
	if length < 1 {
		panic("Trying to split a tree with size < 1")
	}
	uLength := uint(length)
	bitlen := bits.Len(uLength)
	k := int64(1 << uint(bitlen-1))
	if k == length {
		k >>= 1
	}
	return k
}
