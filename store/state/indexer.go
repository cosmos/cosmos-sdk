package state

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

func NewIndexer(m Mapping, enc IntEncoding) Indexer {
	return Indexer{
		m:   m,
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
