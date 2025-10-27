package iavlx

import (
	io "io"

	storetypes "cosmossdk.io/store/types"
)

type Tree struct {
	parent      parentTree
	origRoot    *NodePointer
	root        *NodePointer
	updateBatch KVUpdateBatch
	zeroCopy    bool
}

type parentTree interface {
	getRoot() *NodePointer
	applyChangesToParent(origRoot, newRoot *NodePointer, updateBatch KVUpdateBatch) error
}

func NewTree(parent parentTree, stagedVersion uint32, zeroCopy bool) *Tree {
	root := parent.getRoot()
	return &Tree{
		parent:      parent,
		root:        root,
		origRoot:    root,
		updateBatch: KVUpdateBatch{Version: stagedVersion},
		zeroCopy:    zeroCopy,
	}
}

func (tree *Tree) getRoot() *NodePointer {
	return tree.root
}

func (tree *Tree) applyChangesToParent(origRoot, newRoot *NodePointer, updateBatch KVUpdateBatch) error {
	if tree.root != origRoot {
		panic("cannot apply changes: root has changed")
	}
	tree.root = newRoot
	tree.updateBatch.Updates = append(tree.updateBatch.Updates, updateBatch.Updates...)
	tree.updateBatch.Orphans = append(tree.updateBatch.Orphans, updateBatch.Orphans...)
	return nil
}

func (tree *Tree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (tree *Tree) CacheWrap() storetypes.CacheWrap {
	return NewTree(tree, tree.updateBatch.Version, tree.zeroCopy)
}

func (tree *Tree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO support tracing
	return tree.CacheWrap()
}

func (tree *Tree) Write() {
	if tree.parent == nil {
		panic("cannot write: tree is immutable")
	}
	err := tree.parent.applyChangesToParent(tree.origRoot, tree.root, tree.updateBatch)
	if err != nil {
		panic(err)
	}
	tree.updateBatch.Updates = nil
	tree.updateBatch.Orphans = nil
	tree.root = tree.parent.getRoot()
	tree.origRoot = tree.root
}

func (tree *Tree) Get(key []byte) []byte {
	if tree.root == nil {
		return nil
	}

	root, err := tree.root.Resolve()
	if err != nil {
		panic(err)
	}

	value, _, err := root.Get(key)
	if err != nil {
		panic(err)
	}

	return value
}

func (tree *Tree) Set(key, value []byte) {
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
		panic(err)
	}

	tree.root = newRoot
	tree.updateBatch.Updates = append(tree.updateBatch.Updates, KVUpdate{
		SetNode: leafNode,
	})
	tree.updateBatch.Orphans = append(tree.updateBatch.Orphans, ctx.Orphans)
}

func (tree *Tree) Delete(key []byte) {
	ctx := &MutationContext{Version: tree.updateBatch.Version}
	_, newRoot, _, err := removeRecursive(tree.root, key, ctx)
	if err != nil {
		panic(err)
	}
	tree.root = newRoot
	tree.updateBatch.Updates = append(tree.updateBatch.Updates, KVUpdate{
		DeleteKey: key,
	})
	tree.updateBatch.Orphans = append(tree.updateBatch.Orphans, ctx.Orphans)
}

func (tree *Tree) Has(key []byte) bool {
	// TODO optimize this if possible
	val := tree.Get(key)
	return val != nil
}

func (tree *Tree) Iterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, true, tree.root, tree.zeroCopy)
}

func (tree *Tree) ReverseIterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, false, tree.root, tree.zeroCopy)
}

var (
	_ storetypes.CacheWrap = (*Tree)(nil)
	_ storetypes.KVStore   = (*Tree)(nil)
	_ parentTree           = (*Tree)(nil)
)
