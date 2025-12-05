package internal

import "fmt"

// Node represents a traversable node in the IAVL tree.
type Node interface {
	// ID returns the unique identifier of the node.
	// If the node has not been assigned an ID yet, it returns the zero value of NodeID.
	ID() NodeID

	// IsLeaf indicates whether this node is a leaf node.
	IsLeaf() bool

	// Key returns the key of this node.
	Key() (UnsafeBytes, error)

	// Value returns the value of this node. It is an error to call this method on non-leaf nodes.
	Value() (UnsafeBytes, error)

	// Left returns a pointer to the left child node.
	// If this is called on a leaf node, it returns nil.
	Left() *NodePointer

	// Right returns a pointer to the right child node.
	// If this is called on a leaf node, it returns nil.
	Right() *NodePointer

	// Hash returns the hash of this node.
	// Hash may or may not have been computed yet.
	Hash() UnsafeBytes

	// Height returns the height of the subtree rooted at this node.
	Height() uint8

	// Size returns the number of leaf nodes in the subtree rooted at this node.
	Size() int64

	// Version returns the version at which this node was created.
	Version() uint32

	// Get traverses this subtree to find the value associated with the given key.
	Get(key []byte) (value UnsafeBytes, index int64, err error)

	// MutateBranch creates a mutable copy of this branch node created at the specified version.
	// Since this is an immutable tree, whenever we need to modify a branch node, we should call this method
	// to create a mutable copy of it with its version updated.
	// This method should only be called on branch nodes; calling it on leaf nodes will result in an error.
	MutateBranch(version uint32) (*MemNode, error)

	fmt.Stringer
}

type UnsafeBytes struct {
	bz   []byte
	safe bool
}

func WrapUnsafeBytes(bz []byte) UnsafeBytes {
	return UnsafeBytes{bz: bz, safe: false}
}

func WrapSafeBytes(bz []byte) UnsafeBytes {
	return UnsafeBytes{bz: bz, safe: true}
}

func (ub UnsafeBytes) UnsafeBytes() []byte {
	return ub.bz
}

func (ub UnsafeBytes) SafeCopy() []byte {
	if ub.safe {
		return ub.bz
	}
	if ub.bz == nil {
		return nil
	}
	copied := make([]byte, len(ub.bz))
	copy(copied, ub.bz)
	return copied
}
