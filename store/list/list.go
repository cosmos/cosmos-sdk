package list

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// Key for the length of the list
func LengthKey() []byte {
	return []byte{0x00}
}

// Key for the elements of the list
func ElemKey(index uint64) []byte {
	return append([]byte{0x01}, []byte(fmt.Sprintf("%020d", index))...)
}

// List defines an integer indexable mapper
// It panics when the element type cannot be (un/)marshalled by the codec
type List struct {
	cdc   *codec.Codec
	store types.KVStore
}

// NewList constructs new List
func NewList(cdc *codec.Codec, store types.KVStore) List {
	return List{
		cdc:   cdc,
		store: store,
	}
}

// Len() returns the length of the list
// The length is only increased by Push() and not decreased
// List dosen't check if an index is in bounds
// The user should check Len() before doing any actions
func (m List) Len() (res uint64) {
	bz := m.store.Get(LengthKey())
	if bz == nil {
		return 0
	}

	m.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)
	return
}

// Get() returns the element by its index
func (m List) Get(index uint64, ptr interface{}) error {
	bz := m.store.Get(ElemKey(index))
	return m.cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

// Set() stores the element to the given position
// Setting element out of range will break length counting
// Use Push() instead of Set() to append a new element
func (m List) Set(index uint64, value interface{}) {
	bz := m.cdc.MustMarshalBinaryLengthPrefixed(value)
	m.store.Set(ElemKey(index), bz)
}

// Delete() deletes the element in the given position
// Other elements' indices are preserved after deletion
// Panics when the index is out of range
func (m List) Delete(index uint64) {
	m.store.Delete(ElemKey(index))
}

// Push() inserts the element to the end of the list
// It will increase the length when it is called
func (m List) Push(value interface{}) {
	length := m.Len()
	m.Set(length, value)
	m.store.Set(LengthKey(), m.cdc.MustMarshalBinaryLengthPrefixed(length+1))
}

// Iterate() is used to iterate over all existing elements in the list
// Return true in the continuation to break
// The second element of the continuation will indicate the position of the element
// Using it with Get() will return the same one with the provided element

// CONTRACT: No writes may happen within a domain while iterating over it.
func (m List) Iterate(ptr interface{}, fn func(uint64) bool) {
	iter := types.KVStorePrefixIterator(m.store, []byte{0x01})
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		v := iter.Value()
		m.cdc.MustUnmarshalBinaryLengthPrefixed(v, ptr)

		k := iter.Key()
		s := string(k[len(k)-20:])

		index, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			panic(err)
		}

		if fn(index) {
			break
		}
	}
}
