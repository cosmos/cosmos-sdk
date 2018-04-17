package lib

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

// ListMapper is a Mapper interface that provides list-like functions
// It panics when the element type cannot be (un/)marshalled by the codec

type ListMapper interface {
	// ListMapper dosen't check if an index is in bounds
	// The user should check Len() before doing any actions
	Len(sdk.Context) uint64

	Get(sdk.Context, uint64, interface{}) error

	// Setting element out of range is harmful
	// Use Push() instead of Set() to append a new element
	Set(sdk.Context, uint64, interface{})

	Delete(sdk.Context, uint64)

	Push(sdk.Context, interface{})

	// Iterate*() is used to iterate over all existing elements in the list
	// Return true in the continuation to break

	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	IterateRead(sdk.Context, interface{}, func(sdk.Context, uint64) bool)

	// IterateWrite() is safe to write over the domain
	IterateWrite(sdk.Context, interface{}, func(sdk.Context, uint64) bool)
}

type listMapper struct {
	key    sdk.StoreKey
	cdc    *wire.Codec
	prefix string
}

func NewListMapper(cdc *wire.Codec, key sdk.StoreKey, prefix string) ListMapper {
	return listMapper{
		key:    key,
		cdc:    cdc,
		prefix: prefix,
	}
}

func (lm listMapper) Len(ctx sdk.Context) uint64 {
	store := ctx.KVStore(lm.key)
	bz := store.Get(lm.LengthKey())
	if bz == nil {
		zero, err := lm.cdc.MarshalBinary(0)
		if err != nil {
			panic(err)
		}
		store.Set(lm.LengthKey(), zero)
		return 0
	}
	var res uint64
	if err := lm.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}
	return res
}

func (lm listMapper) Get(ctx sdk.Context, index uint64, ptr interface{}) error {
	store := ctx.KVStore(lm.key)
	bz := store.Get(lm.ElemKey(index))
	return lm.cdc.UnmarshalBinary(bz, ptr)
}

func (lm listMapper) Set(ctx sdk.Context, index uint64, value interface{}) {
	store := ctx.KVStore(lm.key)
	bz, err := lm.cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	store.Set(lm.ElemKey(index), bz)
}

func (lm listMapper) Delete(ctx sdk.Context, index uint64) {
	store := ctx.KVStore(lm.key)
	store.Delete(lm.ElemKey(index))
}

func (lm listMapper) Push(ctx sdk.Context, value interface{}) {
	length := lm.Len(ctx)
	lm.Set(ctx, length, value)

	store := ctx.KVStore(lm.key)
	store.Set(lm.LengthKey(), marshalUint64(lm.cdc, length+1))
}

func (lm listMapper) IterateRead(ctx sdk.Context, ptr interface{}, fn func(sdk.Context, uint64) bool) {
	store := ctx.KVStore(lm.key)
	start, end := subspace([]byte(fmt.Sprintf("%s/elem/", lm.prefix)))
	iter := store.Iterator(start, end)
	for ; iter.Valid(); iter.Next() {
		v := iter.Value()
		if err := lm.cdc.UnmarshalBinary(v, ptr); err != nil {
			panic(err)
		}
		s := strings.Split(string(iter.Key()), "/")
		index, err := strconv.ParseUint(s[len(s)-1], 10, 64)
		if err != nil {
			panic(err)
		}
		if fn(ctx, index) {
			break
		}
	}

	iter.Close()
}

func (lm listMapper) IterateWrite(ctx sdk.Context, ptr interface{}, fn func(sdk.Context, uint64) bool) {
	length := lm.Len(ctx)

	for i := uint64(0); i < length; i++ {
		if err := lm.Get(ctx, i, ptr); err != nil {
			continue
		}
		if fn(ctx, i) {
			break
		}
	}
}

func (lm listMapper) LengthKey() []byte {
	return []byte(fmt.Sprintf("%s/length", lm.prefix))
}

func (lm listMapper) ElemKey(i uint64) []byte {
	return []byte(fmt.Sprintf("%s/elem/%020d", lm.prefix, i))
}

// QueueMapper is a Mapper interface that provides queue-like functions
// It panics when the element type cannot be (un/)marshalled by the codec

type QueueMapper interface {
	Push(sdk.Context, interface{})
	// Popping/Peeking on an empty queue will cause panic
	// The user should check IsEmpty() before doing any actions
	Peek(sdk.Context, interface{}) error
	Pop(sdk.Context)
	IsEmpty(sdk.Context) bool
	// Iterate() removes elements it processed
	// Return true in the continuation to break
	// The interface{} is unmarshalled before the continuation is called
	// Starts from the top(head) of the queue
	Flush(sdk.Context, interface{}, func(sdk.Context) bool)
}

type queueMapper struct {
	key    sdk.StoreKey
	cdc    *wire.Codec
	prefix string
	lm     ListMapper
}

func NewQueueMapper(cdc *wire.Codec, key sdk.StoreKey, prefix string) QueueMapper {
	return queueMapper{
		key:    key,
		cdc:    cdc,
		prefix: prefix,
		lm:     NewListMapper(cdc, key, prefix+"list"),
	}
}

func (qm queueMapper) getTop(store sdk.KVStore) (res uint64) {
	bz := store.Get(qm.TopKey())
	if bz == nil {
		store.Set(qm.TopKey(), marshalUint64(qm.cdc, 0))
		return 0
	}

	if err := qm.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}

	return
}

func (qm queueMapper) setTop(store sdk.KVStore, top uint64) {
	bz := marshalUint64(qm.cdc, top)
	store.Set(qm.TopKey(), bz)
}

func (qm queueMapper) Push(ctx sdk.Context, value interface{}) {
	qm.lm.Push(ctx, value)
}

func (qm queueMapper) Peek(ctx sdk.Context, ptr interface{}) error {
	store := ctx.KVStore(qm.key)
	top := qm.getTop(store)
	return qm.lm.Get(ctx, top, ptr)
}

func (qm queueMapper) Pop(ctx sdk.Context) {
	store := ctx.KVStore(qm.key)
	top := qm.getTop(store)
	qm.lm.Delete(ctx, top)
	qm.setTop(store, top+1)
}

func (qm queueMapper) IsEmpty(ctx sdk.Context) bool {
	store := ctx.KVStore(qm.key)
	top := qm.getTop(store)
	length := qm.lm.Len(ctx)
	return top >= length
}

func (qm queueMapper) Flush(ctx sdk.Context, ptr interface{}, fn func(sdk.Context) bool) {
	store := ctx.KVStore(qm.key)
	top := qm.getTop(store)
	length := qm.lm.Len(ctx)

	var i uint64
	for i = top; i < length; i++ {
		qm.lm.Get(ctx, i, ptr)
		qm.lm.Delete(ctx, i)
		if fn(ctx) {
			break
		}
	}

	qm.setTop(store, i)
}

func (qm queueMapper) TopKey() []byte {
	return []byte(fmt.Sprintf("%s/top", qm.prefix))
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
