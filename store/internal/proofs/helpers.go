package proofs

import (
	"maps"
	"slices"

	cmtprotocrypto "github.com/cometbft/cometbft/api/cometbft/crypto/v1"

	"cosmossdk.io/math/unsafe"
	sdkmaps "cosmossdk.io/store/internal/maps"
)

// SimpleResult contains a merkle.SimpleProof along with all data needed to build the confio/proof
type SimpleResult struct {
	Key      []byte
	Value    []byte
	Proof    *cmtprotocrypto.Proof
	RootHash []byte
}

// GenerateRangeProof creates a merkle tree of specified size and generates a range proof
// for a randomly selected element at the specified location (Left, Right, or Middle).
// Returns a SimpleResult containing the proof data, key, value, and root hash.
func GenerateRangeProof(size int, loc Where) *SimpleResult {
	data := BuildMap(size)
	root, proofs, allkeys := sdkmaps.ProofsFromMap(data)

	key := GetKey(allkeys, loc)
	proof := proofs[key]

	res := &SimpleResult{
		Key:      []byte(key),
		Value:    toValue(key),
		Proof:    proof,
		RootHash: root,
	}
	return res
}

// Where selects a location for a key within a sorted key set
type Where int

const (
	Left   Where = iota // First key in the sorted set
	Right               // Last key in the sorted set
	Middle              // Random key from the middle of the set
)

// SortedKeys returns all keys from the data map in sorted order.
// Useful for deterministic key selection and proof generation.
func SortedKeys(data map[string][]byte) []string {
	return slices.Sorted(maps.Keys(data))
}

// CalcRoot calculates the merkle root hash for the given key-value data.
// Returns only the root hash, discarding proofs and keys for efficiency.
func CalcRoot(data map[string][]byte) []byte {
	root, _, _ := sdkmaps.ProofsFromMap(data)
	return root
}

// GetKey returns a key from the sorted key list based on the specified location.
// Left returns the first key, Right returns the last key, and Middle returns
// a randomly selected key from the middle range (excluding first and last).
func GetKey(allkeys []string, loc Where) string {
	if loc == Left {
		return allkeys[0]
	}
	if loc == Right {
		return allkeys[len(allkeys)-1]
	}
	// select a random index between 1 and allkeys-2
	idx := unsafe.NewRand().Int()%(len(allkeys)-2) + 1
	return allkeys[idx]
}

// GetNonKey returns a key that is guaranteed not to exist in the data set.
// Left returns a key smaller than all existing keys, Right returns a key
// larger than all existing keys, and Middle returns a key that falls
// between existing keys in the sorted order.
func GetNonKey(allkeys []string, loc Where) string {
	if loc == Left {
		return string([]byte{1, 1, 1, 1})
	}
	if loc == Right {
		return string([]byte{0xff, 0xff, 0xff, 0xff})
	}
	// otherwise, next to an existing key (copy before mod)
	key := GetKey(allkeys, loc)
	key = key[:len(key)-2] + string([]byte{255, 255})
	return key
}

// toValue generates a deterministic value for a given key.
// Used for testing purposes to create predictable key-value pairs.
func toValue(key string) []byte {
	return []byte("value_for_" + key)
}

// BuildMap creates a map with random key-value pairs for testing.
// Generates the specified number of entries with random 20-byte keys
// and corresponding deterministic values. Returns the map and a
// sorted list of all generated keys.
func BuildMap(size int) map[string][]byte {
	data := make(map[string][]byte)
	// insert lots of info and store the bytes
	for i := 0; i < size; i++ {
		key := unsafe.Str(20)
		data[key] = toValue(key)
	}
	return data
}
