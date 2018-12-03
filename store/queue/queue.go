package store

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Key for the top element position in the queue
func TopKey() []byte {
	return []byte{0x02}
}

// Queue is a List wrapper that provides queue-like functions
// It panics when the element type cannot be (un/)marshalled by the codec
type Queue struct {
	List List
}

// NewQueue constructs new Queue
func NewQueue(cdc *codec.Codec, store sdk.KVStore) Queue {
	return Queue{NewList(cdc, store)}
}

func (m Queue) getTop() (res uint64) {
	bz := m.List.store.Get(TopKey())
	if bz == nil {
		return 0
	}

	m.List.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)
	return
}

func (m Queue) setTop(top uint64) {
	bz := m.List.cdc.MustMarshalBinaryLengthPrefixed(top)
	m.List.store.Set(TopKey(), bz)
}

// Push() inserts the elements to the rear of the queue
func (m Queue) Push(value interface{}) {
	m.List.Push(value)
}

// Popping/Peeking on an empty queue will cause panic
// The user should check IsEmpty() before doing any actions
// Peek() returns the element at the front of the queue without removing it
func (m Queue) Peek(ptr interface{}) error {
	top := m.getTop()
	return m.List.Get(top, ptr)
}

// Pop() returns the element at the front of the queue and removes it
func (m Queue) Pop() {
	top := m.getTop()
	m.List.Delete(top)
	m.setTop(top + 1)
}

// IsEmpty() checks if the queue is empty
func (m Queue) IsEmpty() bool {
	top := m.getTop()
	length := m.List.Len()
	return top >= length
}

// Flush() removes elements it processed
// Return true in the continuation to break
// The interface{} is unmarshalled before the continuation is called
// Starts from the top(head) of the queue
// CONTRACT: Pop() or Push() should not be performed while flushing
func (m Queue) Flush(ptr interface{}, fn func() bool) {
	top := m.getTop()
	length := m.List.Len()

	var i uint64
	for i = top; i < length; i++ {
		err := m.List.Get(i, ptr)
		if err != nil {
			// TODO: Handle with #870
			panic(err)
		}
		m.List.Delete(i)
		if fn() {
			break
		}
	}
	m.setTop(i)
}
