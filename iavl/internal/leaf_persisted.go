package internal

import (
	"bytes"
	"fmt"
)

type LeafPersisted struct {
	store  *ChangesetReader
	layout *LeafLayout
}

func (node *LeafPersisted) ID() NodeID {
	return node.layout.ID
}

func (node *LeafPersisted) Height() uint8 {
	return 0
}

func (node *LeafPersisted) IsLeaf() bool {
	return true
}

func (node *LeafPersisted) Size() int64 {
	return 1
}

func (node *LeafPersisted) Version() uint32 {
	return uint32(node.layout.Version)
}

func (node *LeafPersisted) Key() (UnsafeBytes, error) {
	return readBlob(node.store, node.layout.KeyOffset)
}

func (node *LeafPersisted) Value() (UnsafeBytes, error) {
	return readBlob(node.store, node.layout.ValueOffset)
}

func (node *LeafPersisted) Left() *NodePointer {
	return nil
}

func (node *LeafPersisted) Right() *NodePointer {
	return nil
}

func (node *LeafPersisted) Hash() UnsafeBytes {
	return WrapUnsafeBytes(node.layout.Hash[:])
}

func (node *LeafPersisted) MutateBranch(uint32) (*MemNode, error) {
	return nil, fmt.Errorf("leaf nodes should not get mutated this way")
}

func (node *LeafPersisted) Get(key []byte) (value UnsafeBytes, index int64, err error) {
	nodeKey, err := node.Key()
	if err != nil {
		return UnsafeBytes{}, 0, err
	}
	switch bytes.Compare(nodeKey.UnsafeBytes(), key) {
	case -1:
		return UnsafeBytes{}, 1, nil
	case 1:
		return UnsafeBytes{}, 0, nil
	default:
		value, err := node.Value()
		if err != nil {
			return UnsafeBytes{}, 0, err
		}
		return value, 0, nil
	}
}

func (node *LeafPersisted) String() string {
	// TODO implement me
	panic("implement me")
}

var _ Node = (*LeafPersisted)(nil)
