package iavlx

import corestore "cosmossdk.io/core/store"

type Tree struct {
	origRoot    *NodePointer
	root        *NodePointer
	updateBatch *KVUpdateBatch
	zeroCopy    bool
}

func NewTree(root *NodePointer, updateBatch *KVUpdateBatch, zeroCopy bool) *Tree {
	return &Tree{origRoot: root, root: root, updateBatch: updateBatch, zeroCopy: zeroCopy}
}

func (tree *Tree) Get(key []byte) ([]byte, error) {
	if tree.root == nil {
		return nil, nil
	}

	root, err := tree.root.Resolve()
	if err != nil {
		return nil, err
	}

	value, _, err := root.Get(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (tree *Tree) Set(key, value []byte) error {
	leafNode := &MemNode{
		height:  0,
		size:    1,
		version: tree.updateBatch.Version,
		key:     key,
		value:   value,
	}
	ctx := &MutationContext{Version: tree.updateBatch.Version}
	newRoot, _, err := setRecursive(tree.root, leafNode, ctx)
	if err != nil {
		return err
	}

	tree.root = newRoot
	tree.updateBatch.Updates = append(tree.updateBatch.Updates, KVUpdate{
		SetNode: leafNode,
	})
	tree.updateBatch.Orphans = append(tree.updateBatch.Orphans, ctx.Orphans)
	return nil
}

func (tree *Tree) Delete(key []byte) error {
	ctx := &MutationContext{Version: tree.updateBatch.Version}
	_, newRoot, _, err := removeRecursive(tree.root, key, ctx)
	if err != nil {
		return err
	}
	tree.root = newRoot
	tree.updateBatch.Updates = append(tree.updateBatch.Updates, KVUpdate{
		DeleteKey: key,
	})
	tree.updateBatch.Orphans = append(tree.updateBatch.Orphans, ctx.Orphans)
	return nil
}

func (tree *Tree) Has(key []byte) (bool, error) {
	// TODO optimize this
	val, err := tree.Get(key)
	if err != nil {
		return false, err
	}
	return val != nil, nil
}

func (tree *Tree) Iterator(start, end []byte) (corestore.Iterator, error) {
	return NewIterator(start, end, true, tree.root, tree.zeroCopy), nil
}

func (tree *Tree) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return NewIterator(start, end, false, tree.root, tree.zeroCopy), nil
}

var _ corestore.KVStore = &Tree{}
