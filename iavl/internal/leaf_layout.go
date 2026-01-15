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
	// NOTE: that a 32-bit offset means that the key data file can be at most 4GB in size.
	// If we want to support larger key/value data files in the future, we can change this to a 40-bit offset.
	// However, this would require growing the size of this struct which would break
	// on-disk compatibility.
	// Such an upgrade could be made by introducing a "wide changeset" format that lives alongside
	// this existing "compact" format.
	KeyOffset Uint40

	// ValueOffset is the offset the value data for this node in the key value data file.
	// The same size considerations apply here as for KeyOffset.
	ValueOffset Uint40

	// Hash is the hash of this leaf node.
	Hash [32]byte
}

func (l LeafLayout) GetNodeID() NodeID {
	return l.ID
}
