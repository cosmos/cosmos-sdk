package internal

import "fmt"

// Node represents a traversable node in the IAVL tree.
type Node interface {
	// ID returns the unique identifier of the node.
	// If the node has not been assigned an ID yet, it returns the zero value of NodeID.
	ID() NodeID

	// IsLeaf indicates whether this node is a leaf node.
	IsLeaf() bool

	// CmpKey compares the given key with the key of this node.
	// It returns:
	//   - a negative integer if this node's key is less than otherKey,
	//   - zero if they are equal,
	//   - a positive integer if this node's key is greater than otherKey.
	// Prefer this method over Key() for comparisons to avoid unnecessary disk reads,
	// because in many cases we can determine the ordering by just comparing key prefixes stored inline
	// in branch nodes.
	CmpKey(otherKey []byte) (int, error)

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
	// If the key is found, value contains the associated value.
	// If the key is not found, value.IsNil() will return true (not an error).
	// The index is the 0-based position where the key exists or would be inserted
	// in sorted order among all leaf keys in this subtree. This is useful for
	// range queries and determining a key's position even when it doesn't exist.
	Get(key []byte) (value UnsafeBytes, index int64, err error)

	// Has traverses this subtree to check if the given key exists.
	// If the key exists, exists is true and index is the 0-based position of the key among all leaf keys in this subtree.
	// If the key does not exist, exists is false and index is the 0-based position where the key would be inserted in sorted order among all leaf keys in this subtree.
	// This method is more efficient than Get when we only need to check for existence and position of the key without needing the value,
	Has(key []byte) (exists bool, index int64, err error)

	// MutateBranch creates a mutable copy of this branch node created at the specified version.
	// Since this is an immutable tree, whenever we need to modify a branch node, we should call this method
	// to create a mutable copy of it with its version updated.
	// This method should only be called on branch nodes; calling it on leaf nodes will result in an error.
	MutateBranch(version uint32) (*MemNode, error)

	fmt.Stringer
}
