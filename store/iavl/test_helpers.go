package iavl

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/libs/rand"
	db "github.com/tendermint/tm-db"
)

// Result is the result of one match
type Result struct {
	Key      []byte
	Value    []byte
	Proof    *iavl.RangeProof
	RootHash []byte
}

// GenerateResult makes a tree of size and returns a range proof for one random element
//
// returns a range proof and the root hash of the tree
func GenerateResult(size int, loc Where) (*Result, error) {
	tree, allkeys, err := BuildTree(size)
	if err != nil {
		return nil, err
	}
	key := GetKey(allkeys, loc)

	value, proof, err := tree.GetWithProof(key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, fmt.Errorf("GetWithProof returned nil value")
	}
	if len(proof.Leaves) != 1 {
		return nil, fmt.Errorf("GetWithProof returned %d leaves", len(proof.Leaves))
	}
	root := tree.Hash()

	res := &Result{
		Key:      key,
		Value:    value,
		Proof:    proof,
		RootHash: root,
	}
	return res, nil
}

// Where selects a location for a key - Left, Right, or Middle
type Where int

const (
	Left Where = iota
	Right
	Middle
)

// GetKey this returns a key, on Left/Right/Middle
func GetKey(allkeys [][]byte, loc Where) []byte {
	if loc == Left {
		return allkeys[0]
	}
	if loc == Right {
		return allkeys[len(allkeys)-1]
	}
	// select a random index between 1 and allkeys-2
	idx := rand.Int()%(len(allkeys)-2) + 1
	return allkeys[idx]
}

// GetNonKey returns a missing key - Left of all, Right of all, or in the Middle
func GetNonKey(allkeys [][]byte, loc Where) []byte {
	if loc == Left {
		return []byte{0, 0, 0, 1}
	}
	if loc == Right {
		return []byte{0xff, 0xff, 0xff, 0xff}
	}
	// otherwise, next to an existing key (copy before mod)
	key := append([]byte{}, GetKey(allkeys, loc)...)
	key[len(key)-2] = 255
	key[len(key)-1] = 255
	return key
}

// BuildTree creates random key/values and stores in tree
// returns a list of all keys in sorted order
func BuildTree(size int) (itree *iavl.ImmutableTree, keys [][]byte, err error) {
	tree, _ := iavl.NewMutableTree(db.NewMemDB(), 0)

	// insert lots of info and store the bytes
	keys = make([][]byte, size)
	for i := 0; i < size; i++ {
		key := rand.Str(20)
		value := "value_for_" + key
		tree.Set([]byte(key), []byte(value))
		keys[i] = []byte(key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i], keys[j]) < 0
	})

	return tree.ImmutableTree, keys, nil
}
