package internal

import (
	"fmt"
	"unsafe"
)

const (
	sizeLeaf = 44
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
	// KeyOffset is the offset the key data for this node in the key value data file.
	KeyOffset uint32
	// Hash is the hash of this leaf node.
	Hash [32]byte
}
