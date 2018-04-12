package stdlib

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

// ListMapper is a Mapper interface that provides list-like functions
// It panics when the element type cannot be (un/)marshalled by the codec

type ListMapper interface {
	// ListMapper dosen't checks index out of range
	// The user should check Len() before doing any actions
	Len(sdk.Context) int64
	Get(sdk.Context, int64, interface{})
	// Setting element out of range is harmful; use Push() when adding new elements
	Set(sdk.Context, int64, interface{})
	Delete(sdk.Context, int64)
	Push(sdk.Context, interface{})
	Iterate(sdk.Context, interface{}, func(sdk.Context, int64) bool)
}

type listMapper struct {
	key    sdk.StoreKey
	cdc    *wire.Codec
	prefix string
	lk     []byte
}

func NewListMapper(cdc *wire.Codec, key sdk.StoreKey, prefix string) ListMapper {
	lk, err := cdc.MarshalBinary(int64(-1))
	if err != nil {
		panic(err)
	}
	return listMapper{
		key:    key,
		cdc:    cdc,
		prefix: prefix,
		lk:     lk,
	}
}

func (lm listMapper) Len(ctx sdk.Context) int64 {
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
	var res int64
	if err := lm.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}
	return res
}

func (lm listMapper) Get(ctx sdk.Context, index int64, ptr interface{}) {
	if index < 0 {
		panic(fmt.Errorf("Invalid index in ListMapper.Get(ctx, %d, ptr)", index))
	}
	store := ctx.KVStore(lm.key)
	bz := store.Get(lm.ElemKey(index))
	if err := lm.cdc.UnmarshalBinary(bz, ptr); err != nil {
		panic(err)
	}
}

func (lm listMapper) Set(ctx sdk.Context, index int64, value interface{}) {
	if index < 0 {
		panic(fmt.Errorf("Invalid index in ListMapper.Set(ctx, %d, value)", index))
	}
	store := ctx.KVStore(lm.key)
	bz, err := lm.cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	store.Set(lm.ElemKey(index), bz)
}

func (lm listMapper) Delete(ctx sdk.Context, index int64) {
	if index < 0 {
		panic(fmt.Errorf("Invalid index in ListMapper.Delete(ctx, %d)", index))
	}
	store := ctx.KVStore(lm.key)
	store.Delete(lm.ElemKey(index))
}

func (lm listMapper) Push(ctx sdk.Context, value interface{}) {
	length := lm.Len(ctx)
	lm.Set(ctx, length, value)

	store := ctx.KVStore(lm.key)
	store.Set(lm.LengthKey(), marshalInt64(lm.cdc, length+1))
}

func (lm listMapper) Iterate(ctx sdk.Context, ptr interface{}, fn func(sdk.Context, int64) bool) {
	length := lm.Len(ctx)
	for i := int64(0); i < length; i++ {
		lm.Get(ctx, i, ptr)
		if fn(ctx, i) {
			break
		}
	}
}

func (lm listMapper) LengthKey() []byte {
	return []byte(fmt.Sprintf("%s/%d", lm.prefix, lm.lk))
}

func (lm listMapper) ElemKey(i int64) []byte {
	return []byte(fmt.Sprintf("%s/%d", lm.prefix, i))
}

// QueueMapper is a Mapper interface that provides queue-like functions
// It panics when the element type cannot be (un/)marshalled by the codec

type QueueMapper interface {
	Push(sdk.Context, interface{})
	// Popping/Peeking on an empty queue will cause panic
	// The user should check IsEmpty() before doing any actions
	Peek(sdk.Context, interface{})
	Pop(sdk.Context)
	IsEmpty(sdk.Context) bool
	// Iterate() removes elements it processed; return true in the continuation to break
	Iterate(sdk.Context, interface{}, func(sdk.Context) bool)
}

type queueMapper struct {
	key    sdk.StoreKey
	cdc    *wire.Codec
	prefix string
	lm     ListMapper
	lk     []byte
	ik     []byte
}

func NewQueueMapper(cdc *wire.Codec, key sdk.StoreKey, prefix string) QueueMapper {
	lk := []byte("list")
	ik := []byte("info")
	return queueMapper{
		key:    key,
		cdc:    cdc,
		prefix: prefix,
		lm:     NewListMapper(cdc, key, prefix+string(lk)),
		lk:     lk,
		ik:     ik,
	}
}

type queueInfo struct {
	// begin <= elems < end
	Begin int64
	End   int64
}

func (info queueInfo) validateBasic() error {
	if info.End < info.Begin || info.Begin < 0 || info.End < 0 {
		return fmt.Errorf("Invalid queue information: {Begin: %d, End: %d}", info.Begin, info.End)
	}
	return nil
}

func (info queueInfo) isEmpty() bool {
	return info.Begin == info.End
}

func (qm queueMapper) getQueueInfo(store sdk.KVStore) queueInfo {
	bz := store.Get(qm.InfoKey())
	if bz == nil {
		store.Set(qm.InfoKey(), marshalQueueInfo(qm.cdc, queueInfo{0, 0}))
		return queueInfo{0, 0}
	}
	var info queueInfo
	if err := qm.cdc.UnmarshalBinary(bz, &info); err != nil {
		panic(err)
	}
	if err := info.validateBasic(); err != nil {
		panic(err)
	}
	return info
}

func (qm queueMapper) setQueueInfo(store sdk.KVStore, info queueInfo) {
	bz, err := qm.cdc.MarshalBinary(info)
	if err != nil {
		panic(err)
	}
	store.Set(qm.InfoKey(), bz)
}

func (qm queueMapper) Push(ctx sdk.Context, value interface{}) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)

	qm.lm.Set(ctx, info.End, value)

	info.End++
	qm.setQueueInfo(store, info)
}

func (qm queueMapper) Peek(ctx sdk.Context, ptr interface{}) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)
	qm.lm.Get(ctx, info.Begin, ptr)
}

func (qm queueMapper) Pop(ctx sdk.Context) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)
	qm.lm.Delete(ctx, info.Begin)
	info.Begin++
	qm.setQueueInfo(store, info)
}

func (qm queueMapper) IsEmpty(ctx sdk.Context) bool {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)
	return info.isEmpty()
}

func (qm queueMapper) Iterate(ctx sdk.Context, ptr interface{}, fn func(sdk.Context) bool) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)

	var i int64
	for i = info.Begin; i < info.End; i++ {
		qm.lm.Get(ctx, i, ptr)
		key := marshalInt64(qm.cdc, i)
		store.Delete(key)
		if fn(ctx) {
			break
		}
	}

	info.Begin = i
	qm.setQueueInfo(store, info)
}

func (qm queueMapper) InfoKey() []byte {
	return []byte(fmt.Sprintf("%s/%s", qm.prefix, qm.ik))
}

func marshalQueueInfo(cdc *wire.Codec, info queueInfo) []byte {
	bz, err := cdc.MarshalBinary(info)
	if err != nil {
		panic(err)
	}
	return bz
}

func marshalInt64(cdc *wire.Codec, i int64) []byte {
	bz, err := cdc.MarshalBinary(i)
	if err != nil {
		panic(err)
	}
	return bz
}
