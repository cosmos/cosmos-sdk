package internal

import (
	"fmt"
	"unsafe"
)

const (
	sizeLeaf       = 56
	LeafLayoutSize = sizeLeaf
)

func init() {
	// Verify the size of LeafLayout is what we expect it to be at runtime.
	if unsafe.Sizeof(LeafLayout{}) != sizeLeaf {
		panic(fmt.Sprintf("invalid LeafLayout size: got %d, want %d", unsafe.Sizeof(LeafLayout{}), sizeLeaf))
	}
}

// LeafLayout is the on-disk layout of a leaf node.
// NOTE: changes to this struct will affect on-disk compatibility.
type LeafLayout struct {
	// ID is the NodeID of this leaf node.
	ID NodeID

	Version uint32

	// KeyOffset is the offset the key data for this node in the key value data file.
	KeyOffset Uint40

	// ValueOffset is the offset the value data for this node in the key value data file.
	ValueOffset Uint40

	flags uint8

	// Hash is the hash of this leaf node.
	Hash [32]byte
}

func (l LeafLayout) GetNodeID() NodeID {
	return l.ID
}

const (
	leafFlagKeyInKVData   = 0x01
	leafFlagValueInKVData = 0x02
)

// KeyInKVData returns true if the key for this leaf node is stored in kv.dat, false if it's stored in wal.log.
func (l LeafLayout) KeyInKVData() bool {
	return l.flags&leafFlagKeyInKVData != 0
}

// ValueInKVData returns true if the value for this leaf node is stored in kv.dat, false if it's stored in wal.log.
func (l LeafLayout) ValueInKVData() bool {
	return l.flags&leafFlagValueInKVData != 0
}

// SetKeyInKVData sets whether the key for this leaf node is stored in kv.dat or wal.log.
func (l *LeafLayout) SetKeyInKVData(inKVData bool) {
	if inKVData {
		l.flags |= leafFlagKeyInKVData
	} else {
		l.flags &^= leafFlagKeyInKVData
	}
}

// SetValueInKVData sets whether the value for this leaf node is stored in kv.dat or wal.log.
func (l *LeafLayout) SetValueInKVData(inKVData bool) {
	if inKVData {
		l.flags |= leafFlagValueInKVData
	} else {
		l.flags &^= leafFlagValueInKVData
	}
}
