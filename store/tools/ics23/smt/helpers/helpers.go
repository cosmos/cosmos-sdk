/*
Package helpers contains functions to build sample data for tests/testgen

In it's own package to avoid poluting the godoc for ics23-smt
*/
package helpers

import (
	"bytes"
	"crypto/sha256"
	"math/rand"
	"sort"

	"github.com/lazyledger/smt"

	tmproofs "github.com/cosmos/cosmos-sdk/store/internal/proofs"
)

// PreimageMap maps each tree path back to its preimage
// needed because SparseMerkleTree methods take preimage as arg and hash internally
type PreimageMap struct {
	paths []preimageMapping
	keys  [][]byte
	// known non-keys at left and rightmost positions
	nonKeys []preimageMapping
}
type preimageMapping struct {
	path   [32]byte
	keyIdx int // index of preimage in keys list
}

// BuildTree creates random key/values and stores in tree
// returns a list of all keys in sorted order
func BuildTree(size int) (*smt.SparseMerkleTree, *PreimageMap, error) {
	nodes, values := smt.NewSimpleMap(), smt.NewSimpleMap()
	tree := smt.NewSparseMerkleTree(nodes, values, sha256.New())

	// insert lots of info and store the bytes
	keys := make([][]byte, size+2)
	for i := 0; i < len(keys); i++ {
		key := randStr(20)

		value := "value_for_" + key
		_, err := tree.Update([]byte(key), []byte(value))
		if err != nil {
			return nil, nil, err
		}
		keys[i] = []byte(key)
	}

	var paths []preimageMapping
	for i, key := range keys {
		paths = append(paths, preimageMapping{sha256.Sum256(key), i})
	}
	sort.Slice(paths, func(i, j int) bool {
		return bytes.Compare(paths[i].path[:], paths[j].path[:]) < 0
	})

	// now, find the edge paths and remove them from the tree
	leftmost, rightmost := paths[0], paths[len(paths)-1]
	_, err := tree.Delete(keys[leftmost.keyIdx])
	if err != nil {
		return nil, nil, err
	}
	_, err = tree.Delete(keys[rightmost.keyIdx])
	if err != nil {
		return nil, nil, err
	}

	pim := PreimageMap{
		keys:    keys,
		paths:   paths[1 : len(paths)-1],
		nonKeys: []preimageMapping{leftmost, rightmost},
	}
	return tree, &pim, nil
}

// FindPath returns the closest index to path in paths, and whether it's a match.
// If not found, the returned index is where the path would be.
func (pim PreimageMap) FindPath(path [32]byte) (int, bool) {
	var mid int
	from, to := 0, len(pim.paths)-1
	for from <= to {
		mid = (from + to) / 2
		switch bytes.Compare(pim.paths[mid].path[:], path[:]) {
		case -1:
			from = mid + 1
		case 1:
			to = mid - 1
		default:
			return mid, true
		}
	}
	return from, false
}

// Len returns the number of mapped paths.
func (pim PreimageMap) Len() int { return len(pim.paths) }

// KeyFor returns the preimage (key) for given path index.
func (pim PreimageMap) KeyFor(pathIx int) []byte {
	return pim.keys[pim.paths[pathIx].keyIdx]
}

// GetKey this returns a key, on Left/Right/Middle
func (pim PreimageMap) GetKey(loc tmproofs.Where) []byte {
	if loc == tmproofs.Left {
		return pim.KeyFor(0)
	}
	if loc == tmproofs.Right {
		return pim.KeyFor(len(pim.paths) - 1)
	}
	// select a random index between 1 and len-2
	idx := rand.Int()%(len(pim.paths)-2) + 1
	return pim.KeyFor(idx)
}

// GetNonKey returns a missing key - Left of all, Right of all, or in the Middle
func (pim PreimageMap) GetNonKey(loc tmproofs.Where) []byte {
	if loc == tmproofs.Left {
		return pim.keys[pim.nonKeys[0].keyIdx]
	}
	if loc == tmproofs.Right {
		return pim.keys[pim.nonKeys[1].keyIdx]
	}
	// otherwise, next to an existing key (copy before mod)
	key := append([]byte{}, pim.GetKey(tmproofs.Middle)...)
	key[len(key)-2] = 255
	key[len(key)-1] = 255
	return key
}
