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

func (node *LeafPersisted) CmpKey(otherKey []byte) (int, error) {
	key, err := node.Key()
	if err != nil {
		return 0, err
	}
	return bytes.Compare(key.UnsafeBytes(), otherKey), nil
}

func (node *LeafPersisted) Key() (UnsafeBytes, error) {
	return readLeafBlob(node.store, node.layout.KeyOffset)
}

func (node *LeafPersisted) Value() (UnsafeBytes, error) {
	return readLeafBlob(node.store, node.layout.ValueOffset)
}

func readLeafBlob(rdr *ChangesetReader, offset Uint40) (UnsafeBytes, error) {
	// leaf data is stored either in the WAL OR KV data depending on whether this is a
	// compaction changeset or not
	kvData := rdr.WALData()
	if kvData == nil {
		kvData = rdr.KVData()
	}
	bz, err := kvData.UnsafeReadBlob(int(offset.ToUint64()))
	if err != nil {
		return UnsafeBytes{}, err
	}
	return WrapUnsafeBytes(bz), nil
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
