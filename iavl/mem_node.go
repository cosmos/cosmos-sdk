package iavlx

import (
	"bytes"
	"fmt"
)

type MemNode struct {
	height    uint8
	size      int64
	version   uint32
	key       []byte
	value     []byte
	left      *NodePointer
	right     *NodePointer
	hash      []byte
	nodeId    NodeID // ID of this node, 0 if not yet assigned
	keyOffset uint32
}

func (node *MemNode) ID() NodeID {
	return node.nodeId
}

func (node *MemNode) Height() uint8 {
	return node.height
}

func (node *MemNode) Size() int64 {
	return node.size
}

func (node *MemNode) Version() uint32 {
	return node.version
}

func (node *MemNode) Key() ([]byte, error) {
	return node.key, nil
}

func (node *MemNode) Value() ([]byte, error) {
	return node.value, nil
}

func (node *MemNode) Left() *NodePointer {
	return node.left
}

func (node *MemNode) Right() *NodePointer {
	return node.right
}

func (node *MemNode) Hash() []byte {
	return node.hash
}

func (node *MemNode) SafeHash() []byte {
	// TODO what needs to be safe??
	return node.hash
}

func (node *MemNode) MutateBranch(version uint32) (*MemNode, error) {
	n := *node
	n.version = version
	n.hash = nil
	return &n, nil
}

func (node *MemNode) Get(key []byte) (value []byte, index int64, err error) {
	if node.IsLeaf() {
		switch bytes.Compare(node.key, key) {
		case -1:
			return nil, 1, nil
		case 1:
			return nil, 0, nil
		default:
			return node.value, 0, nil
		}
	}

	if bytes.Compare(key, node.key) < 0 {
		leftNode, err := node.left.Resolve()
		if err != nil {
			return nil, 0, err
		}

		return leftNode.Get(key)
	}

	rightNode, err := node.right.Resolve()
	if err != nil {
		return nil, 0, err
	}

	value, index, err = rightNode.Get(key)
	if err != nil {
		return nil, 0, err
	}

	index += node.size - rightNode.Size()
	return value, index, nil
}

func (node *MemNode) IsLeaf() bool {
	return node.height == 0
}

func (node *MemNode) String() string {
	if node.IsLeaf() {
		return fmt.Sprintf("MemNode{key:%x, version:%d, size:%d, value:%x}", node.key, node.version, node.size, node.value)
	} else {
		return fmt.Sprintf("MemNode{key:%x, version:%d, size:%d, height:%d, left:%s, right:%s}", node.key, node.version, node.size, node.height, node.left, node.right)
	}
}

var _ Node = &MemNode{}
