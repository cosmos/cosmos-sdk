package internal

import (
	"fmt"
	"unsafe"
)

const (
	sizeBranch = 80
)

func init() {
	// Verify the size of BranchLayout is what we expect it to be at runtime.
	if unsafe.Sizeof(BranchLayout{}) != sizeBranch {
		panic(fmt.Sprintf("invalid BranchLayout size: got %d, want %d", unsafe.Sizeof(BranchLayout{}), sizeBranch))
	}
}

// BranchLayout is the on-disk layout of a branch node.
// NOTE: changes to this struct will affect on-disk compatibility.
type BranchLayout struct {
	// ID is the NodeID of this branch node.
	ID NodeID

	Version uint32

	// Left is the NodeID of the left child node.
	Left NodeID

	// Right is the NodeID of the right child node.
	Right NodeID

	// NOTE: Left and right offsets are included for performance and take up an extra 8 bytes of storage for each branch node.
	// In an alternate design we stored only NodeID or offset for left and right depending on whether they are local
	// to this changeset or in a different changeset.
	// This saved 8 bytes of storage per branch node but made the implementation significantly more complex.
	// For now, we are including both the left and right IDs and offsets, but if storage space becomes a problem
	// we can revisit the earlier design and have an 8-byte NodeIDOrOffset type for Left and Right.

	// LeftOffset is the 1-based offset of the left child node if it is in this changeset, 0 otherwise.
	// The Left NodeID will indicate whether this is a branch or leaf node.
	LeftOffset uint32

	// RightOffset is the 1-based offset of the right child node if it is in this changeset, 0 otherwise.
	// The Right NodeID will indicate whether this is a branch or leaf node.
	RightOffset uint32

	// KeyOffset is the offset the key data for this node in the key value data file.
	// NOTE: that a 32-bit offset means that the key data file can be at most 4GB in size.
	// This doesn't limit the size of the overall tree, it just limits the size of individual key/value data files.
	// If we want to support larger key/value data files in the future, we can change this to a 40-bit offset,
	// and an additional byte of padding is already reserved below for this purpose.
	KeyOffset uint32

	// Height is the height of the subtree rooted at this branch node.
	Height uint8

	// NOTE: there are two bytes of padding here that could be used for something else in the future if needed
	// such as an extra byte to allow for 40-bit key offsets.

	// Size is the number of leaf nodes in the subtree rooted at this branch node.
	Size Uint40

	// Hash is the hash of this branch node.
	Hash [32]byte
}

func (b BranchLayout) GetNodeID() NodeID {
	return b.ID
}
