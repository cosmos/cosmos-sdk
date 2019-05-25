package mapping

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

type IndexEncoding byte

const (
	DecIndexerEnc IndexEncoding = iota
	HexIndexerEnc
	BinIndexerEnc
)

type Indexer struct {
	m Mapping

	enc IndexEncoding
}

func NewIndexer(base Base, prefix []byte, enc IndexEncoding) Indexer {
	return Indexer{
		m:   NewMapping(base, prefix),
		enc: enc,
	}
}

// Identical length independent from the index, ensure ordering
func EncodeIndex(index uint64, enc IndexEncoding) (res []byte) {
	switch enc {
	case DecIndexerEnc:
		return []byte(fmt.Sprintf("%020d", index))
	case HexIndexerEnc:
		return []byte(fmt.Sprintf("%020x", index))
	case BinIndexerEnc:
		res = make([]byte, 8)
		binary.BigEndian.PutUint64(res, index)
		return
	default:
		panic("invalid IndexEncoding")
	}
}

func DecodeIndex(bz []byte, enc IndexEncoding) (res uint64, err error) {
	switch enc {
	case DecIndexerEnc:
		return strconv.ParseUint(string(bz), 10, 64)
	case HexIndexerEnc:
		return strconv.ParseUint(string(bz), 16, 64)
	case BinIndexerEnc:
		return binary.BigEndian.Uint64(bz), nil
	default:
		panic("invalid IndexEncoding")
	}
}

func (ix Indexer) Value(index uint64) Value {
	return ix.m.Value(EncodeIndex(index, ix.enc))
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
		m: ix.m.Range(EncodeIndex(start, ix.enc), EncodeIndex(end, ix.enc)),

		enc: ix.enc,
	}
}

func (ix Indexer) IterateAscending(ctx Context, ptr interface{}, fn func(uint64) bool) {
	ix.m.Iterate(ctx, ptr, func(bz []byte) bool {
		key, err := DecodeIndex(bz, ix.enc)
		if err != nil {
			panic(err)
		}
		return fn(key)
	})
}

func (ix Indexer) IterateDescending(ctx Context, ptr interface{}, fn func(uint64) bool) {
	ix.m.ReverseIterate(ctx, ptr, func(bz []byte) bool {
		key, err := DecodeIndex(bz, ix.enc)
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
		key, err := DecodeIndex(keybz, ix.enc)
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
		key, err := DecodeIndex(keybz, ix.enc)
		if err != nil {
			return key, false
		}
	}
	return
}

func (ix Indexer) Key(index uint64) []byte {
	return ix.m.Key(EncodeIndex(index, ix.enc))
}
