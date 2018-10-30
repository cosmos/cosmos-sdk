package queue

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/store/list"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Prefix for the PrefixStore of the queue
func ListKey() []byte {
	return []byte{0x00}
}

// Key for the top element position in the queue
func TopKey() []byte {
	return []byte{0x01}
}

// Queue is a List wrapper that provides queue-like functions
// It panics when the element type cannot be (un/)marshalled by the codec
type Queue struct {
	cdc   *codec.Codec
	store types.KVStore

	List list.List
}

// New constructs new Queue
func New(cdc *codec.Codec, store types.KVStore) Queue {
	return Queue{cdc, store, list.New(cdc, prefix.NewStore(store, ListKey()))}
}

func (m Queue) getTop() (res uint64) {
	bz := m.store.Get(TopKey())
	if bz == nil {
		return 0
	}

	m.cdc.MustUnmarshalBinary(bz, &res)
	return
}

func (m Queue) setTop(top uint64) {
	bz := m.cdc.MustMarshalBinary(top)
	m.store.Set(TopKey(), bz)
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
