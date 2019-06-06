package mapping

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

type IntEncoding byte

const (
	Dec IntEncoding = iota
	Hex
	Bin
)

type Indexer struct {
	m Mapping

	enc IntEncoding
}

func NewIndexer(base Base, prefix []byte, enc IntEncoding) Indexer {
	return Indexer{
		m:   NewMapping(base, prefix),
		enc: enc,
	}
}

// Identical length independent from the index, ensure ordering
func EncodeInt(index uint64, enc IntEncoding) (res []byte) {
	switch enc {
	case Dec:
		return []byte(fmt.Sprintf("%020d", index))
	case Hex:
		return []byte(fmt.Sprintf("%020x", index))
	case Bin:
		res = make([]byte, 8)
		binary.BigEndian.PutUint64(res, index)
		return
	default:
		panic("invalid IntEncoding")
	}
}

func DecodeInt(bz []byte, enc IntEncoding) (res uint64, err error) {
	switch enc {
	case Dec:
		return strconv.ParseUint(string(bz), 10, 64)
	case Hex:
		return strconv.ParseUint(string(bz), 16, 64)
	case Bin:
		return binary.BigEndian.Uint64(bz), nil
	default:
		panic("invalid IntEncoding")
	}
}

func (ix Indexer) Value(index uint64) Value {
	return ix.m.Value(EncodeInt(index, ix.enc))
}

func (ix Indexer) Get(ctx Context, index uint64, ptr interface{}) {
	ix.Value(index).Get(ctx, ptr)
}

func (ix Indexer) GetIfExists(ctx Context, index uint64, ptr interface{}) {
	ix.Value(index).GetIfExists(ctx, ptr)
}

func (ix Indexer) GetSafe(ctx Context, index uint64, ptr interface{}) error {
	return ix.Value(index).GetSafe(ctx, ptr)
}

func (ix Indexer) Set(ctx Context, index uint64, o interface{}) {
	ix.Value(index).Set(ctx, o)
}

func (ix Indexer) Has(ctx Context, index uint64) bool {
	return ix.Value(index).Exists(ctx)
}

func (ix Indexer) Delete(ctx Context, index uint64) {
	ix.Value(index).Delete(ctx)
}

func (ix Indexer) IsEmpty(ctx Context) bool {
	return ix.m.IsEmpty(ctx)
}

func (ix Indexer) Prefix(prefix []byte) Indexer {
	return Indexer{
		m: ix.m.Prefix(prefix),

		enc: ix.enc,
	}
}

func (ix Indexer) Range(start, end uint64) Indexer {
	return Indexer{
		m: ix.m.Range(EncodeInt(start, ix.enc), EncodeInt(end, ix.enc)),

		enc: ix.enc,
	}
}

func (ix Indexer) IterateAscending(ctx Context, ptr interface{}, fn func(uint64) bool) {
	ix.m.Iterate(ctx, ptr, func(bz []byte) bool {
		key, err := DecodeInt(bz, ix.enc)
		if err != nil {
			panic(err)
		}
		return fn(key)
	})
}

func (ix Indexer) IterateDescending(ctx Context, ptr interface{}, fn func(uint64) bool) {
	ix.m.ReverseIterate(ctx, ptr, func(bz []byte) bool {
		key, err := DecodeInt(bz, ix.enc)
		if err != nil {
			panic(err)
		}
		return fn(key)
	})
}

func (ix Indexer) First(ctx Context, ptr interface{}) (key uint64, ok bool) {
	keybz, ok := ix.m.First(ctx, ptr)
	if !ok {
		return
	}
	if len(keybz) != 0 {
		key, err := DecodeInt(keybz, ix.enc)
		if err != nil {
			return key, false
		}
	}
	return
}

func (ix Indexer) Last(ctx Context, ptr interface{}) (key uint64, ok bool) {
	keybz, ok := ix.m.Last(ctx, ptr)
	if !ok {
		return
	}
	if len(keybz) != 0 {
		key, err := DecodeInt(keybz, ix.enc)
		if err != nil {
			return key, false
		}
	}
	return
}

/*
func (ix Indexer) Key(index uint64) []byte {
	return ix.m.Key(EncodeInt(index, ix.enc))
}
*/
