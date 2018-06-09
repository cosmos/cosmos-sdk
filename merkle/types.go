package merkle

import (
	"io"

	amino "github.com/tendermint/go-amino"
)

// Tree is a Merkle tree interface.
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

// Hasher represents a hashable piece of data which can be hashed in the Tree.
type Hasher interface {
	Hash() []byte
}

//-----------------------------------------------------------------------

// Uvarint length prefixed byteslice
func encodeByteSlice(w io.Writer, bz []byte) (err error) {
	return amino.EncodeByteSlice(w, bz)
}
