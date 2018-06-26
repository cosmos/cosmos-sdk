package lib

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

// Linear defines a primitive mapper type
type Linear struct {
	cdc   *wire.Codec
	store sdk.KVStore
	keys  *LinearKeys
}

// LinearKeys defines keysions for the key bytes
type LinearKeys struct {
	LengthKey []byte
	ElemKey   []byte
	TopKey    []byte
}

// Should never be modified
var cachedDefaultLinearKeys = DefaultLinearKeys()

// DefaultLinearKeys returns the default setting of LinearOption
func DefaultLinearKeys() *LinearKeys {
	keys := LinearKeys{
		LengthKey: []byte{0x00},
		ElemKey:   []byte{0x01},
		TopKey:    []byte{0x02},
	}
	return &keys
}

// NewLinear constructs new Linear
func NewLinear(cdc *wire.Codec, store sdk.KVStore, keys *LinearKeys) Linear {
	if keys == nil {
		keys = cachedDefaultLinearKeys
	}
	if keys.LengthKey == nil || keys.ElemKey == nil || keys.TopKey == nil {
		panic("Invalid LinearKeys")
	}
	return Linear{
		cdc:   cdc,
		store: store,
		keys:  keys,
	}
}

// List is a Linear interface that provides list-like functions
// It panics when the element type cannot be (un/)marshalled by the codec
type List interface {

	// Len() returns the length of the list
	// The length is only increased by Push() and not decreased
	// List dosen't check if an index is in bounds
	// The user should check Len() before doing any actions
	Len() uint64

	// Get() returns the element by its index
	Get(uint64, interface{}) error

	// Set() stores the element to the given position
	// Setting element out of range will break length counting
	// Use Push() instead of Set() to append a new element
	Set(uint64, interface{})

	// Delete() deletes the element in the given position
	// Other elements' indices are preserved after deletion
	// Panics when the index is out of range
	Delete(uint64)

	// Push() inserts the element to the end of the list
	// It will increase the length when it is called
	Push(interface{})

	// Iterate*() is used to iterate over all existing elements in the list
	// Return true in the continuation to break
	// The second element of the continuation will indicate the position of the element
	// Using it with Get() will return the same one with the provided element

	// CONTRACT: No writes may happen within a domain while iterating over it.
	Iterate(interface{}, func(uint64) bool)
}

// NewList constructs new List
func NewList(cdc *wire.Codec, store sdk.KVStore, keys *LinearKeys) List {
	return NewLinear(cdc, store, keys)
}

// Key for the length of the list
func (m Linear) LengthKey() []byte {
	return m.keys.LengthKey
}

// Key for the elements of the list
func (m Linear) ElemKey(index uint64) []byte {
	return append(m.keys.ElemKey, []byte(fmt.Sprintf("%020d", index))...)
}

// Len implements List
func (m Linear) Len() (res uint64) {
	bz := m.store.Get(m.LengthKey())
	if bz == nil {
		return 0
	}
	m.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// Get implements List
func (m Linear) Get(index uint64, ptr interface{}) error {
	bz := m.store.Get(m.ElemKey(index))
	return m.cdc.UnmarshalBinary(bz, ptr)
}

// Set implements List
func (m Linear) Set(index uint64, value interface{}) {
	bz := m.cdc.MustMarshalBinary(value)
	m.store.Set(m.ElemKey(index), bz)
}

// Delete implements List
func (m Linear) Delete(index uint64) {
	m.store.Delete(m.ElemKey(index))
}

// Push implements List
func (m Linear) Push(value interface{}) {
	length := m.Len()
	m.Set(length, value)
	m.store.Set(m.LengthKey(), m.cdc.MustMarshalBinary(length+1))
}

// IterateRead implements List
func (m Linear) Iterate(ptr interface{}, fn func(uint64) bool) {
	iter := sdk.KVStorePrefixIterator(m.store, []byte{0x01})
	for ; iter.Valid(); iter.Next() {
		v := iter.Value()
		m.cdc.MustUnmarshalBinary(v, ptr)
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

	iter.Close()
}

// Queue is a Linear interface that provides queue-like functions
// It panics when the element type cannot be (un/)marshalled by the codec
type Queue interface {
	// Push() inserts the elements to the rear of the queue
	Push(interface{})

	// Popping/Peeking on an empty queue will cause panic
	// The user should check IsEmpty() before doing any actions

	// Peek() returns the element at the front of the queue without removing it
	Peek(interface{}) error

	// Pop() returns the element at the front of the queue and removes it
	Pop()

	// IsEmpty() checks if the queue is empty
	IsEmpty() bool

	// Flush() removes elements it processed
	// Return true in the continuation to break
	// The interface{} is unmarshalled before the continuation is called
	// Starts from the top(head) of the queue
	// CONTRACT: Pop() or Push() should not be performed while flushing
	Flush(interface{}, func() bool)
}

// NewQueue constructs new Queue
func NewQueue(cdc *wire.Codec, store sdk.KVStore, keys *LinearKeys) Queue {
	return NewLinear(cdc, store, keys)
}

// Key for the top element position in the queue
func (m Linear) TopKey() []byte {
	return m.keys.TopKey
}

func (m Linear) getTop() (res uint64) {
	bz := m.store.Get(m.TopKey())
	if bz == nil {
		return 0
	}

	m.cdc.MustUnmarshalBinary(bz, &res)
	return
}

func (m Linear) setTop(top uint64) {
	bz := m.cdc.MustMarshalBinary(top)
	m.store.Set(m.TopKey(), bz)
}

// Peek implements Queue
func (m Linear) Peek(ptr interface{}) error {
	top := m.getTop()
	return m.Get(top, ptr)
}

// Pop implements Queue
func (m Linear) Pop() {
	top := m.getTop()
	m.Delete(top)
	m.setTop(top + 1)
}

// IsEmpty implements Queue
func (m Linear) IsEmpty() bool {
	top := m.getTop()
	length := m.Len()
	return top >= length
}

// Flush implements Queue
func (m Linear) Flush(ptr interface{}, fn func() bool) {
	top := m.getTop()
	length := m.Len()

	var i uint64
	for i = top; i < length; i++ {
		m.Get(i, ptr)
		m.Delete(i)
		if fn() {
			break
		}
	}
	m.setTop(i)
}

func subspace(prefix []byte) (start, end []byte) {
	end = make([]byte, len(prefix))
	copy(end, prefix)
	end[len(end)-1]++
	return prefix, end
}
