package internal

import (
	"bytes"
	"fmt"
)

type BranchPersisted struct {
	store  *ChangesetReader
	layout *BranchLayout
}

func (node *BranchPersisted) Left() *NodePointer {
	return &NodePointer{
		changeset: node.store.Changeset(),
		fileIdx:   node.layout.LeftOffset,
		id:        node.layout.Left,
	}
}

func (node *BranchPersisted) Right() *NodePointer {
	return &NodePointer{
		changeset: node.store.Changeset(),
		fileIdx:   node.layout.RightOffset,
		id:        node.layout.Right,
	}
}

func (node *BranchPersisted) ID() NodeID {
	return node.layout.ID
}

func (node *BranchPersisted) Height() uint8 {
	return node.layout.Height
}

func (node *BranchPersisted) IsLeaf() bool {
	return false
}

func (node *BranchPersisted) Size() int64 {
	return int64(node.layout.Size.ToUint64())
}

func (node *BranchPersisted) Version() uint32 {
	return node.layout.Version
}

func (node *BranchPersisted) Key() (UnsafeBytes, error) {
	// the key data may be stored either in the WAL OR KV data depending on the key info flag
	kvData := node.store.WALData()
	if kvData == nil || node.layout.KeyInfo.InKVData() {
		kvData = node.store.KVData()
	}
	bz, err := kvData.UnsafeReadBlob(int(node.layout.KeyOffset.ToUint64()))
	if err != nil {
		return UnsafeBytes{}, err
	}
	return WrapUnsafeBytes(bz), nil
}

func (node *BranchPersisted) Value() (UnsafeBytes, error) {
	return UnsafeBytes{}, fmt.Errorf("branch nodes do not have values")
}

func (node *BranchPersisted) Hash() UnsafeBytes {
	return WrapUnsafeBytes(node.layout.Hash[:])
}

func (node *BranchPersisted) MutateBranch(version uint32) (*MemNode, error) {
	key, err := node.Key()
	if err != nil {
		return nil, err
	}

	memNode := &MemNode{
		height:  node.Height(),
		size:    node.Size(),
		version: version,
		key:     key.SafeCopy(),
		left:    node.Left(),
		right:   node.Right(),
	}
	return memNode, err
}

func (node *BranchPersisted) Get(key []byte) (value UnsafeBytes, index int64, err error) {
	nodeKey, err := node.Key()
	if err != nil {
		return UnsafeBytes{}, 0, err
	}

	if bytes.Compare(key, nodeKey.UnsafeBytes()) < 0 {
		leftNode, leftPin, err := node.Left().Resolve()
		defer leftPin.Unpin()
		if err != nil {
			return UnsafeBytes{}, 0, err
		}

		return leftNode.Get(key)
	}

	rightNode, rightPin, err := node.Right().Resolve()
	defer rightPin.Unpin()
	if err != nil {
		return UnsafeBytes{}, 0, err
	}

	value, index, err = rightNode.Get(key)
	if err != nil {
		return UnsafeBytes{}, 0, err
	}

	index += node.Size() - rightNode.Size()
	return value, index, nil
}

func (node *BranchPersisted) String() string {
	// TODO implement me
	panic("implement me")
}

var _ Node = (*BranchPersisted)(nil)
