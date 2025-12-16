package internal

import "fmt"

// NodeID is a stable identifier for a node in the IAVL tree.
// A NodeID allows for a 32-bit version and a 31-bit index within that version,
// with 1 bit used to indicate whether the node is a leaf or branch.
// A 32-bit version should allow for 136 years of 1-second blocks.
// If block production significantly speeds up, we can increase the width of the version field in the future.
// This sort of change can be done without any major on-disk migration because we can simply create a "wide changeset"
// format that lives alongside the existing "compact" format.
// Because the cost of migration is low, we have decided to keep things simple and compact for now.
type NodeID struct {
	// Version is the version of the tree at which this node was created.
	Version uint32

	// FlagIndex indicates whether this is a branch or leaf node and stores its index in the tree.
	FlagIndex NodeFlagIndex
}

// NodeFlagIndex is the index of an IAVL node in the tree plus a flag indicating whether this is a branch or leaf node.
// For leaf nodes, the index value is the 1-based in-order index of the leaf node with reference to other leaf nodes in this version.
// For branch nodes, the index value is the 1-based post-order traversal index of the node within this version.
// Bit 31 indicates whether this is a branch or leaf node (0 for branch, 1 for leaf).
type NodeFlagIndex uint32

// NewNodeID creates a new NodeID.
func NewNodeID(isLeaf bool, version, index uint32) NodeID {
	return NodeID{
		Version:   version,
		FlagIndex: NewNodeFlagIndex(isLeaf, index),
	}
}

// IsLeaf returns true if the node is a leaf node.
func (id NodeID) IsLeaf() bool {
	return id.FlagIndex.IsLeaf()
}

// IsEmpty returns true if the NodeID is the zero value.
func (id NodeID) IsEmpty() bool {
	return id.Version == 0 && id.FlagIndex == 0
}

// String returns a string representation of the NodeID.
func (id NodeID) String() string {
	return fmt.Sprintf("NodeID{leaf:%t, version:%d, index:%d}", id.IsLeaf(), id.Version, id.FlagIndex.Index())
}

// NewNodeFlagIndex creates a new NodeFlagIndex.
func NewNodeFlagIndex(isLeaf bool, index uint32) NodeFlagIndex {
	idx := NodeFlagIndex(index)
	if isLeaf {
		idx |= 1 << 31
	}
	return idx
}

// IsLeaf returns true if the node is a leaf node.
func (index NodeFlagIndex) IsLeaf() bool {
	return index&(1<<31) != 0
}

// Index returns the index of the node in the tree.
func (index NodeFlagIndex) Index() uint32 {
	return uint32(index) & 0x7FFFFFFF
}
