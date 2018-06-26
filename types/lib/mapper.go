package lib

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

// Mapper defines a primitive mapper type
type Mapper struct {
	key    sdk.StoreKey
	cdc    *wire.Codec
	prefix string
}

// ListMapper is a Mapper interface that provides list-like functions
// It panics when the element type cannot be (un/)marshalled by the codec
type ListMapper interface {

	// Len() returns the length of the list
	// The length is only increased by Push() and not decreased
	// ListMapper dosen't check if an index is in bounds
	// The user should check Len() before doing any actions
	Len(sdk.Context) uint64

	// Get() returns the element by its index
	Get(sdk.Context, uint64, interface{}) error

	// Set() stores the element to the given position
	// Setting element out of range will break length counting
	// Use Push() instead of Set() to append a new element
	Set(sdk.Context, uint64, interface{})

	// Delete() deletes the element in the given position
	// Other elements' indices are preserved after deletion
	// Panics when the index is out of range
	Delete(sdk.Context, uint64)

	// Push() inserts the element to the end of the list
	// It will increase the length when it is called
	Push(sdk.Context, interface{})

	// Iterate*() is used to iterate over all existing elements in the list
	// Return true in the continuation to break
	// The second element of the continuation will indicate the position of the element
	// Using it with Get() will return the same one with the provided element

	// CONTRACT: No writes may happen within a domain while iterating over it.
	IterateRead(sdk.Context, interface{}, func(sdk.Context, uint64) bool)

	// IterateWrite() is safe to write over the domain
	IterateWrite(sdk.Context, interface{}, func(sdk.Context, uint64) bool)

	// Key for the length of the list
	LengthKey() []byte

	// Key for getting elements
	ElemKey(uint64) []byte
}

// NewListMapper constructs new ListMapper
func NewListMapper(cdc *wire.Codec, key sdk.StoreKey, prefix string) ListMapper {
	return Mapper{
		key:    key,
		cdc:    cdc,
		prefix: prefix,
	}
}

// Len implements ListMapper
func (m Mapper) Len(ctx sdk.Context) uint64 {
	store := ctx.KVStore(m.key)
	bz := store.Get(m.LengthKey())
	if bz == nil {
		zero, err := m.cdc.MarshalBinary(0)
		if err != nil {
			panic(err)
		}
		store.Set(m.LengthKey(), zero)
		return 0
	}
	var res uint64
	if err := m.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}
	return res
}

// Get implements ListMapper
func (m Mapper) Get(ctx sdk.Context, index uint64, ptr interface{}) error {
	store := ctx.KVStore(m.key)
	bz := store.Get(m.ElemKey(index))
	return m.cdc.UnmarshalBinary(bz, ptr)
}

// Set implements ListMapper
func (m Mapper) Set(ctx sdk.Context, index uint64, value interface{}) {
	store := ctx.KVStore(m.key)
	bz, err := m.cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	store.Set(m.ElemKey(index), bz)
}

// Delete implements ListMapper
func (m Mapper) Delete(ctx sdk.Context, index uint64) {
	store := ctx.KVStore(m.key)
	store.Delete(m.ElemKey(index))
}

// Push implements ListMapper
func (m Mapper) Push(ctx sdk.Context, value interface{}) {
	length := m.Len(ctx)
	m.Set(ctx, length, value)

	store := ctx.KVStore(m.key)
	store.Set(m.LengthKey(), marshalUint64(m.cdc, length+1))
}

// IterateRead implements ListMapper
func (m Mapper) IterateRead(ctx sdk.Context, ptr interface{}, fn func(sdk.Context, uint64) bool) {
	store := ctx.KVStore(m.key)
	start, end := subspace([]byte(fmt.Sprintf("%s/elem/", m.prefix)))
	iter := store.Iterator(start, end)
	for ; iter.Valid(); iter.Next() {
		v := iter.Value()
		if err := m.cdc.UnmarshalBinary(v, ptr); err != nil {
			panic(err)
		}
		s := string(iter.Key()[len(m.prefix)+6:])
		index, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			panic(err)
		}
		if fn(ctx, index) {
			break
		}
	}

	iter.Close()
}

// IterateWrite implements ListMapper
func (m Mapper) IterateWrite(ctx sdk.Context, ptr interface{}, fn func(sdk.Context, uint64) bool) {
	length := m.Len(ctx)

	for i := uint64(0); i < length; i++ {
		if err := m.Get(ctx, i, ptr); err != nil {
			continue
		}
		if fn(ctx, i) {
			break
		}
	}
}

// LengthKey implements ListMapper
func (m Mapper) LengthKey() []byte {
	return []byte(fmt.Sprintf("%s/length", m.prefix))
}

// ElemKey implements ListMapper
func (m Mapper) ElemKey(i uint64) []byte {
	return []byte(fmt.Sprintf("%s/elem/%020d", m.prefix, i))
}

// QueueMapper is a Mapper interface that provides queue-like functions
// It panics when the element type cannot be (un/)marshalled by the codec
type QueueMapper interface {
	// Push() inserts the elements to the rear of the queue
	Push(sdk.Context, interface{})

	// Popping/Peeking on an empty queue will cause panic
	// The user should check IsEmpty() before doing any actions

	// Peek() returns the element at the front of the queue without removing it
	Peek(sdk.Context, interface{}) error

	// Pop() returns the element at the front of the queue and removes it
	Pop(sdk.Context)

	// IsEmpty() checks if the queue is empty
	IsEmpty(sdk.Context) bool

	// Flush() removes elements it processed
	// Return true in the continuation to break
	// The interface{} is unmarshalled before the continuation is called
	// Starts from the top(head) of the queue
	// CONTRACT: Pop() or Push() should not be performed while flushing
	Flush(sdk.Context, interface{}, func(sdk.Context) bool)

	// Key for the index of top element
	TopKey() []byte
}

// NewQueueMapper constructs new QueueMapper
func NewQueueMapper(cdc *wire.Codec, key sdk.StoreKey, prefix string) QueueMapper {
	return Mapper{
		key:    key,
		cdc:    cdc,
		prefix: prefix,
	}
}

func (m Mapper) getTop(store sdk.KVStore) (res uint64) {
	bz := store.Get(m.TopKey())
	if bz == nil {
		store.Set(m.TopKey(), marshalUint64(m.cdc, 0))
		return 0
	}

	if err := m.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}

	return
}

func (m Mapper) setTop(store sdk.KVStore, top uint64) {
	bz := marshalUint64(m.cdc, top)
	store.Set(m.TopKey(), bz)
}

// Peek implements QueueMapper
func (m Mapper) Peek(ctx sdk.Context, ptr interface{}) error {
	store := ctx.KVStore(m.key)
	top := m.getTop(store)
	return m.Get(ctx, top, ptr)
}

// Pop implements QueueMapper
func (m Mapper) Pop(ctx sdk.Context) {
	store := ctx.KVStore(m.key)
	top := m.getTop(store)
	m.Delete(ctx, top)
	m.setTop(store, top+1)
}

// IsEmpty implements QueueMapper
func (m Mapper) IsEmpty(ctx sdk.Context) bool {
	store := ctx.KVStore(m.key)
	top := m.getTop(store)
	length := m.Len(ctx)
	return top >= length
}

// Flush implements QueueMapper
func (m Mapper) Flush(ctx sdk.Context, ptr interface{}, fn func(sdk.Context) bool) {
	store := ctx.KVStore(m.key)
	top := m.getTop(store)
	length := m.Len(ctx)

	var i uint64
	for i = top; i < length; i++ {
		m.Get(ctx, i, ptr)
		m.Delete(ctx, i)
		if fn(ctx) {
			break
		}
	}

	m.setTop(store, i)
}

// TopKey implements QueueMapper
func (m Mapper) TopKey() []byte {
	return []byte(fmt.Sprintf("%s/top", m.prefix))
}

func marshalUint64(cdc *wire.Codec, i uint64) []byte {
	bz, err := cdc.MarshalBinary(i)
	if err != nil {
		panic(err)
	}
	return bz
}

func subspace(prefix []byte) (start, end []byte) {
	end = make([]byte, len(prefix))
	copy(end, prefix)
	end[len(end)-1]++
	return prefix, end
}
