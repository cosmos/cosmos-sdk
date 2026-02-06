package internal

import (
	"fmt"
	"unsafe"
)

const (
	sizeLeaf = 56
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

func (l LeafLayout) KeyInKVData() bool {
	return l.flags&leafFlagKeyInKVData != 0
}

func (l LeafLayout) ValueInKVData() bool {
	return l.flags&leafFlagValueInKVData != 0
}

func (l *LeafLayout) SetKeyInKVData(inKVData bool) {
	if inKVData {
		l.flags |= leafFlagKeyInKVData
	} else {
		l.flags &^= leafFlagKeyInKVData
	}
}

func (l *LeafLayout) SetValueInKVData(inKVData bool) {
	if inKVData {
		l.flags |= leafFlagValueInKVData
	} else {
		l.flags &^= leafFlagValueInKVData
	}
}
