package types

import (
	"errors"

	wire "github.com/cosmos/cosmos-sdk/wire"
)

type ListMapper interface { // Solidity list like structure
	Len(Context) int64
	Get(Context, int64, interface{})
	Set(Context, int64, interface{})
	Push(Context, interface{})
	Iterate(Context, interface{}, func(Context, int64))
}

type listMapper struct {
	key StoreKey
	cdc *wire.Codec
	lk  []byte
}

func NewListMapper(cdc *wire.Codec, key StoreKey) ListMapper {
	lk, err := cdc.MarshalBinary(int64(-1))
	if err != nil {
		panic(err)
	}
	return listMapper{
		key: key,
		cdc: cdc,
		lk:  lk,
	}
}

func (lm listMapper) Len(ctx Context) int64 {
	store := ctx.KVStore(lm.key)
	bz := store.Get(lm.lk)
	if bz == nil {
		zero, err := lm.cdc.MarshalBinary(0)
		if err != nil {
			panic(err)
		}
		store.Set(lm.lk, zero)
		return 0
	}
	var res int64
	if err := lm.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}
	return res
}

func (lm listMapper) Get(ctx Context, index int64, ptr interface{}) {
	if index < 0 {
		panic(errors.New(""))
	}
	store := ctx.KVStore(lm.key)
	bz := store.Get(marshalInt64(lm.cdc, index))
	if err := lm.cdc.UnmarshalBinary(bz, ptr); err != nil {
		panic(err)
	}
}

func (lm listMapper) Set(ctx Context, index int64, value interface{}) {
	if index < 0 {
		panic(errors.New(""))
	}
	store := ctx.KVStore(lm.key)
	bz, err := lm.cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	store.Set(marshalInt64(lm.cdc, index), bz)
}

func (lm listMapper) Push(ctx Context, value interface{}) {
	length := lm.Len(ctx)
	lm.Set(ctx, length, value)

	store := ctx.KVStore(lm.key)
	store.Set(lm.lk, marshalInt64(lm.cdc, length+1))
}

func (lm listMapper) Iterate(ctx Context, ptr interface{}, fn func(Context, int64)) {
	length := lm.Len(ctx)
	for i := int64(0); i < length; i++ {
		lm.Get(ctx, i, ptr)
		fn(ctx, i)
	}
}

type QueueMapper interface {
	Push(Context, interface{})
	Peek(Context, interface{})
	Pop(Context)
	IsEmpty(Context) bool
	Iterate(Context, interface{}, func(Context) bool)
}

type queueMapper struct {
	key StoreKey
	cdc *wire.Codec
	ik  []byte
}

func NewQueueMapper(cdc *wire.Codec, key StoreKey) QueueMapper {
	ik, err := cdc.MarshalBinary(int64(-1))
	if err != nil {
		panic(err)
	}
	return queueMapper{
		key: key,
		cdc: cdc,
		ik:  ik,
	}
}

type queueInfo struct {
	// begin <= elems < end
	Begin int64
	End   int64
}

func (info queueInfo) validateBasic() error {
	if info.End < info.Begin || info.Begin < 0 || info.End < 0 {
		return errors.New("")
	}
	return nil
}

func (info queueInfo) isEmpty() bool {
	return info.Begin == info.End
}

func (qm queueMapper) getQueueInfo(store KVStore) queueInfo {
	bz := store.Get(qm.ik)
	if bz == nil {
		store.Set(qm.ik, marshalQueueInfo(qm.cdc, queueInfo{0, 0}))
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

func (qm queueMapper) setQueueInfo(store KVStore, info queueInfo) {
	bz, err := qm.cdc.MarshalBinary(info)
	if err != nil {
		panic(err)
	}
	store.Set(qm.ik, bz)
}

func (qm queueMapper) Push(ctx Context, value interface{}) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)

	bz, err := qm.cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	store.Set(marshalInt64(qm.cdc, info.End), bz)

	info.End++
	qm.setQueueInfo(store, info)
}

func (qm queueMapper) Peek(ctx Context, ptr interface{}) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)
	bz := store.Get(marshalInt64(qm.cdc, info.Begin))
	if err := qm.cdc.UnmarshalBinary(bz, ptr); err != nil {
		panic(err)
	}
}

func (qm queueMapper) Pop(ctx Context) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)
	store.Delete(marshalInt64(qm.cdc, info.Begin))
	info.Begin++
	qm.setQueueInfo(store, info)
}

func (qm queueMapper) IsEmpty(ctx Context) bool {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)
	return info.isEmpty()
}

func (qm queueMapper) Iterate(ctx Context, ptr interface{}, fn func(Context) bool) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)

	var i int64
	for i = info.Begin; i < info.End; i++ {
		key := marshalInt64(qm.cdc, i)
		bz := store.Get(key)
		if err := qm.cdc.UnmarshalBinary(bz, ptr); err != nil {
			panic(err)
		}
		store.Delete(key)
		if fn(ctx) {
			break
		}
	}

	info.Begin = i
	qm.setQueueInfo(store, info)
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
