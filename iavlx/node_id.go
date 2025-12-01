package iavlx

import "fmt"

// NodeID is a stable identifier for a node in the IAVL tree.
type NodeID struct {
	Version   uint32
	FlagIndex NodeFlagIndex
}

// NodeFlagIndex is the index of an IAVL node in the tree plus a flag indicating whether this is a branch or leaf node.
// For leaf nodes, the index value is the 1-based in-order index of the leaf node with reference to other leaf nodes in this version.
// For branch nodes, the index value is the 1-based post-order traversal index of the node within this version.
// Bit 31 indicates whether this is a branch or leaf node (0 for branch, 1 for leaf).
type NodeFlagIndex uint32

func NewNodeID(isLeaf bool, version, index uint32) NodeID {
	return NodeID{
		Version:   version,
		FlagIndex: NewNodeFlagIndex(isLeaf, index),
	}
}

func (id NodeID) IsLeaf() bool {
	return id.FlagIndex.IsLeaf()
}

func NewNodeFlagIndex(isLeaf bool, index uint32) NodeFlagIndex {
	idx := NodeFlagIndex(index)
	if isLeaf {
		idx |= 1 << 31
	}
	return idx
}

func (index NodeFlagIndex) IsLeaf() bool {
	return index&(1<<31) != 0
}

func (index NodeFlagIndex) Index() uint32 {
	return uint32(index) & 0x7FFFFFFF
}

func (id NodeID) String() string {
	return fmt.Sprintf("NodeID{leaf:%t, version:%d, index:%d}", id.IsLeaf(), id.Version, id.FlagIndex.Index())
}
