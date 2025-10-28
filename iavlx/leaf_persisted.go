package iavlx

import (
	"bytes"
	"fmt"
)

type LeafPersisted struct {
	store   *Changeset
	selfIdx uint32
	layout  LeafLayout
	// TODO do these need to be atomic?
	key, value []byte
}

func (node *LeafPersisted) ID() NodeID {
	return node.layout.Id
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
	return uint32(node.layout.Id.Version())
}

func (node *LeafPersisted) Key() ([]byte, error) {
	err := node.getKV()
	if err != nil {
		return nil, err
	}
	return node.key, nil
}

func (node *LeafPersisted) Value() ([]byte, error) {
	err := node.getKV()
	if err != nil {
		return nil, err
	}
	return node.value, nil
}

func (node *LeafPersisted) getKV() error {
	if node.key != nil {
		return nil
	}
	k, v, err := node.store.ReadKV(node.layout.Id, node.layout.KeyOffset)
	if err != nil {
		return err
	}
	node.key = k
	node.value = v
	return nil
}

func (node *LeafPersisted) Left() *NodePointer {
	return nil
}

func (node *LeafPersisted) Right() *NodePointer {
	return nil
}

func (node *LeafPersisted) Hash() []byte {
	return node.layout.Hash[:]
}

func (node *LeafPersisted) SafeHash() []byte {
	// TODO how do we make this safe?
	return node.layout.Hash[:]
}

func (node *LeafPersisted) MutateBranch(uint32) (*MemNode, error) {
	return nil, fmt.Errorf("leaf nodes should not get mutated this way")
}

func (node *LeafPersisted) Get(key []byte) (value []byte, index int64, err error) {
	nodeKey, err := node.Key()
	if err != nil {
		return nil, 0, err
	}
	switch bytes.Compare(nodeKey, key) {
	case -1:
		return nil, 1, nil
	case 1:
		return nil, 0, nil
	default:
		value, err := node.Value()
		if err != nil {
			return nil, 0, err
		}
		return value, 0, nil
	}
}

func (node *LeafPersisted) String() string {
	// TODO implement me
	panic("implement me")
}

var _ Node = (*LeafPersisted)(nil)
