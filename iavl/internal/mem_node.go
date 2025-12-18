package internal

import (
	"bytes"
	"fmt"
)

// MemNode represents an in-memory node that has recently been created and may or may not have
// been serialized to disk yet.
type MemNode struct {
	height      uint8
	version     uint32
	size        int64
	key         []byte
	value       []byte
	left        *NodePointer
	right       *NodePointer
	hash        []byte
	nodeId      NodeID // ID of this node, 0 if not yet assigned
	keyOffset   uint32
	valueOffset uint32
}

var _ Node = (*MemNode)(nil)

// ID implements the Node interface.
func (node *MemNode) ID() NodeID {
	return node.nodeId
}

// Height implements the Node interface.
func (node *MemNode) Height() uint8 {
	return node.height
}

// Size implements the Node interface.
func (node *MemNode) Size() int64 {
	return node.size
}

// Version implements the Node interface.
func (node *MemNode) Version() uint32 {
	return node.version
}

// Key implements the Node interface.
func (node *MemNode) Key() (UnsafeBytes, error) {
	return WrapSafeBytes(node.key), nil
}

// Value implements the Node interface.
func (node *MemNode) Value() (UnsafeBytes, error) {
	return WrapSafeBytes(node.value), nil
}

// Left implements the Node interface.
func (node *MemNode) Left() *NodePointer {
	return node.left
}

// Right implements the Node interface.
func (node *MemNode) Right() *NodePointer {
	return node.right
}

// Hash implements the Node interface.
func (node *MemNode) Hash() UnsafeBytes {
	return WrapSafeBytes(node.hash)
}

// MutateBranch implements the Node interface.
func (node *MemNode) MutateBranch(version uint32) (*MemNode, error) {
	n := *node
	n.version = version
	n.hash = nil
	return &n, nil
}

// Get implements the Node interface.
func (node *MemNode) Get(key []byte) (value UnsafeBytes, index int64, err error) {
	if node.IsLeaf() {
		switch bytes.Compare(node.key, key) {
		case -1:
			return UnsafeBytes{}, 1, nil
		case 1:
			return UnsafeBytes{}, 0, nil
		default:
			return WrapSafeBytes(node.value), 0, nil
		}
	}

	if bytes.Compare(key, node.key) < 0 {
		leftNode, pin, err := node.left.Resolve()
		defer pin.Unpin()
		if err != nil {
			return UnsafeBytes{}, 0, err
		}

		return leftNode.Get(key)
	}

	rightNode, pin, err := node.right.Resolve()
	defer pin.Unpin()
	if err != nil {
		return UnsafeBytes{}, 0, err
	}

	value, index, err = rightNode.Get(key)
	if err != nil {
		return UnsafeBytes{}, 0, err
	}

	index += node.size - rightNode.Size()
	return value, index, nil
}

// IsLeaf implements the Node interface.
func (node *MemNode) IsLeaf() bool {
	return node.height == 0
}

// String implements the fmt.Stringer interface.
func (node *MemNode) String() string {
	if node.IsLeaf() {
		return fmt.Sprintf("MemNode{key:%x, version:%d, size:%d, value:%x}", node.key, node.version, node.size, node.value)
	} else {
		return fmt.Sprintf("MemNode{key:%x, version:%d, size:%d, height:%d, left:%s, right:%s}", node.key, node.version, node.size, node.height, node.left, node.right)
	}
}
