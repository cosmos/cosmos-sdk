package iavlx

import "fmt"

// bit 63 indicates whether this is a node ID (0) or relative pointer (1)
// a valid NodeID should always have bit 63 as 0
// bit 62 indicates whether this is a leaf (1) or branch (0)
// bits 61-23 (39 bits) are for version
// bits 22-0 (23 bits) are for index
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
		id |= 1 << 62
	}
	id |= (version & 0x7FFFFFFFFF) << 23
	id |= uint64(index & 0x7FFFFF)
	return NodeID(id)
}

func (id NodeID) IsLeaf() bool {
	// check if second highest bit is set
	return id&(1<<62) != 0
}

func (id NodeID) Version() uint64 {
	return (uint64(id) >> 23) & 0x7FFFFF
}

func (id NodeID) Index() uint32 {
	return uint32(id & 0x7FFFFF)
}

func (id NodeID) String() string {
	return fmt.Sprintf("NodeID{leaf:%t, version:%d, index:%d}", id.IsLeaf(), id.Version(), id.Index())
}

// bit 63 indicates whether this is a node ID (0) or relative pointer (1)
type NodeRef uint64

func (ref NodeRef) IsRelativePointer() bool {
	return ref&(1<<63) != 0
}

func (ref NodeRef) IsNodeID() bool {
	return ref&(1<<63) == 0
}

func (ref NodeRef) IsLeaf() bool {
	return ref&(1<<62) != 0
}

func (ref NodeRef) AsNodeID() NodeID {
	return NodeID(ref)
}

func (ref NodeRef) AsRelativePointer() NodeRelativePointer {
	return NodeRelativePointer(ref &^ (1 << 63))
}

func (ref NodeRef) String() string {
	if ref.IsNodeID() {
		return fmt.Sprintf("NodeRef(%s)", ref.AsNodeID())
	} else {
		return fmt.Sprintf("NodeRef(%s)", ref.AsRelativePointer())
	}
}

// bit 63 indicates whether this is a node ID (0) or relative pointer (1)
// a valid NodeRelativePointer should always have bit 63 as 0
// bit 62 indicates whether this is a leaf (1) or branch (0)
// bits 61-0 (62 bits) are for signed offset
type NodeRelativePointer uint64

func NewNodeRelativePointer(isLeaf bool, offset int64) NodeRelativePointer {
	// check offset fits in 61 bits signed
	if offset < -0x1FFFFFFFFFFFFFFF || offset > 0x1FFFFFFFFFFFFFFF {
		panic("offset too large for NodeRelativePointer")
	}
	var ptr uint64
	ptr |= 1 << 63 // set bit 63 to indicate relative pointer
	if isLeaf {
		ptr |= 1 << 62
	}
	// Store absolute value of offset in bits 60-0
	// Use bit 61 as sign bit (1 = negative)
	if offset < 0 {
		ptr |= 1 << 61                              // set sign bit
		ptr |= uint64(-offset) & 0x1FFFFFFFFFFFFFFF // store absolute value in lower 61 bits
	} else {
		ptr |= uint64(offset) & 0x1FFFFFFFFFFFFFFF // store value in lower 61 bits
	}
	return NodeRelativePointer(ptr)
}

func (ptr NodeRelativePointer) IsLeaf() bool {
	// check if second highest bit is set
	return ptr&(1<<62) != 0
}

func (ptr NodeRelativePointer) Offset() int64 {
	// Extract the absolute value from lower 61 bits
	offset := int64(ptr & 0x1FFFFFFFFFFFFFFF)
	// if bit 61 is set, it's negative
	if ptr&(1<<61) != 0 {
		offset = -offset
	}
	return offset
}

func (ptr NodeRelativePointer) String() string {
	return fmt.Sprintf("NodeRelativePointer{leaf:%t, offset:%d}", ptr.IsLeaf(), ptr.Offset())
}
