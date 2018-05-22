package merkle

import (
	"encoding/binary"
	"io"
)

type Tree interface {
	Size() (size int)
	Height() (height int8)
	Has(key []byte) (has bool)
	Proof(key []byte) (value []byte, proof []byte, exists bool) // TODO make it return an index
	Get(key []byte) (index int, value []byte, exists bool)
	GetByIndex(index int) (key []byte, value []byte)
	Set(key []byte, value []byte) (updated bool)
	Remove(key []byte) (value []byte, removed bool)
	HashWithCount() (hash []byte, count int)
	Hash() (hash []byte)
	Save() (hash []byte)
	Load(hash []byte)
	Copy() Tree
	Iterate(func(key []byte, value []byte) (stop bool)) (stopped bool)
	IterateRange(start []byte, end []byte, ascending bool, fx func(key []byte, value []byte) (stop bool)) (stopped bool)
}

type Hasher interface {
	Hash() []byte
}

//-----------------------------------------------------------------------
// NOTE: these are duplicated from go-amino so we dont need go-amino as a dep

func encodeByteSlice(w io.Writer, bz []byte) (err error) {
	err = encodeUvarint(w, uint64(len(bz)))
	if err != nil {
		return
	}
	_, err = w.Write(bz)
	return
}

func encodeUvarint(w io.Writer, i uint64) (err error) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], i)
	_, err = w.Write(buf[0:n])
	return
}
