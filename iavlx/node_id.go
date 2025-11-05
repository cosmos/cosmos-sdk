package iavlx

import "fmt"

// NodeID is a stable identifier for a node in the IAVL tree.
// Bit 63 indicates whether this is a leaf (1) or branch (0)
// Bits 62-23 (40 bits) are for version.
// Bits 22-0 (23 bits) are for index.
// Valid index values start from 1. A zero index value may be used to indicate a null node.
type NodeID uint64

func NewNodeID(isLeaf bool, version uint64, index uint32) NodeID {
	// check 40 bits for version and 23 bits for index
	if version >= 0x10000000000 {
		panic("version too large for NodeID")
	}
	if index >= 0x800000 {
		panic("index too large for NodeID")
	}
	var id uint64
	if isLeaf {
		id |= 1 << 63
	}
	id |= (version & 0xFFFFFFFFFF) << 23
	id |= uint64(index & 0x7FFFFF)
	return NodeID(id)
}

func (id NodeID) IsLeaf() bool {
	// check if highest bit is set
	return id&(1<<63) != 0
}

func (id NodeID) Version() uint64 {
	return (uint64(id) >> 23) & 0xFFFFFFFFFF
}

func (id NodeID) Index() uint32 {
	return uint32(id & 0x7FFFFF)
}

func (id NodeID) String() string {
	return fmt.Sprintf("NodeID{leaf:%t, version:%d, index:%d}", id.IsLeaf(), id.Version(), id.Index())
}
