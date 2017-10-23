package state

import (
	"bytes"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk"
	wire "github.com/tendermint/go-wire"
)

// SetKey returns the key to get all members of this set
func SetKey() []byte {
	return keys
}

// Set allows us to add arbitrary k-v pairs, check existence,
// as well as iterate through the set (always in key order)
//
// If we had full access to the IAVL tree, this would be completely
// trivial and redundant
type Set struct {
	store sdk.KVStore
	keys  KeyList
}

var _ sdk.KVStore = &Set{}

// NewSet loads or initializes a span of keys
func NewSet(store sdk.KVStore) *Set {
	s := &Set{store: store}
	s.loadKeys()
	return s
}

// Set puts a value at a given height.
// If the value is nil, or an empty slice, remove the key from the list
func (s *Set) Set(key []byte, value []byte) {
	s.store.Set(MakeBKey(key), value)
	if len(value) > 0 {
		s.addKey(key)
	} else {
		s.removeKey(key)
	}
	s.storeKeys()
}

// Get returns the element with a key if it exists
func (s *Set) Get(key []byte) []byte {
	return s.store.Get(MakeBKey(key))
}

// Remove deletes this key from the set (same as setting value = nil)
func (s *Set) Remove(key []byte) {
	s.store.Set(key, nil)
}

// Exists checks for the existence of the key in the set
func (s *Set) Exists(key []byte) bool {
	return len(s.Get(key)) > 0
}

// Size returns how many elements are in the set
func (s *Set) Size() int {
	return len(s.keys)
}

// List returns all keys in the set
// It makes a copy, so we don't modify this in place
func (s *Set) List() (keys KeyList) {
	out := make([][]byte, len(s.keys))
	for i := range s.keys {
		out[i] = append([]byte(nil), s.keys[i]...)
	}
	return out
}

// addKey inserts this key, maintaining sorted order, no duplicates
func (s *Set) addKey(key []byte) {
	for i, k := range s.keys {
		cmp := bytes.Compare(k, key)
		// don't add duplicates
		if cmp == 0 {
			return
		}
		// insert before the first key greater than input
		if cmp > 0 {
			// https://github.com/golang/go/wiki/SliceTricks
			s.keys = append(s.keys, nil)
			copy(s.keys[i+1:], s.keys[i:])
			s.keys[i] = key
			return
		}
	}
	// if it is higher than all (or empty keys), append
	s.keys = append(s.keys, key)
}

// removeKey removes this key if it is present, maintaining sorted order
func (s *Set) removeKey(key []byte) {
	for i, k := range s.keys {
		cmp := bytes.Compare(k, key)
		// if there is a match, remove
		if cmp == 0 {
			s.keys = append(s.keys[:i], s.keys[i+1:]...)
			return
		}
		// if we has the proper location, without finding it, abort
		if cmp > 0 {
			return
		}
	}
}

func (s *Set) loadKeys() {
	b := s.store.Get(keys)
	if b == nil {
		return
	}
	err := wire.ReadBinaryBytes(b, &s.keys)
	// hahaha... just like i love to hate :)
	if err != nil {
		panic(err)
	}
}

func (s *Set) storeKeys() {
	b := wire.BinaryBytes(s.keys)
	s.store.Set(keys, b)
}

// MakeBKey prefixes the byte slice for the storage key
func MakeBKey(key []byte) []byte {
	return append(dataKey, key...)
}

// KeyList is a sortable list of byte slices
type KeyList [][]byte

//nolint
func (kl KeyList) Len() int           { return len(kl) }
func (kl KeyList) Less(i, j int) bool { return bytes.Compare(kl[i], kl[j]) < 0 }
func (kl KeyList) Swap(i, j int)      { kl[i], kl[j] = kl[j], kl[i] }

var _ sort.Interface = KeyList{}

// Equals checks for if the two lists have the same content...
// needed as == doesn't work for slices of slices
func (kl KeyList) Equals(kl2 KeyList) bool {
	if len(kl) != len(kl2) {
		return false
	}
	for i := range kl {
		if !bytes.Equal(kl[i], kl2[i]) {
			return false
		}
	}
	return true
}
