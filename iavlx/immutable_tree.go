package iavlx

import (
	io "io"

	storetypes "cosmossdk.io/store/types"
)

type ImmutableTree struct {
	root *NodePointer
}

func NewImmutableTree(root *NodePointer) *ImmutableTree {
	return &ImmutableTree{
		root: root,
	}
}

func (tree *ImmutableTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (tree *ImmutableTree) CacheWrap() storetypes.CacheWrap {
	return NewCacheTree(tree)
}

func (tree *ImmutableTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO support tracing
	return tree.CacheWrap()
}

func (tree *ImmutableTree) Get(key []byte) []byte {
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

func (tree *ImmutableTree) Set(key, value []byte) {
	panic("cannot set in immutable tree")
}

func (tree *ImmutableTree) Delete(key []byte) {
	panic("cannot delete from immutable tree")
}

func (tree *ImmutableTree) Has(key []byte) bool {
	val := tree.Get(key)
	return val != nil
}

func (tree *ImmutableTree) Iterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, true, tree.root, true)
}

func (tree *ImmutableTree) ReverseIterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, false, tree.root, true)
}

var (
	_ storetypes.KVStore = (*ImmutableTree)(nil)
)
