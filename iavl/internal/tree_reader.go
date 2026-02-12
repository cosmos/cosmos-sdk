package internal

import (
	"io"

	"cosmossdk.io/store/cachekv"
	storetypes "cosmossdk.io/store/types"
)

type TreeReader struct {
	version uint32
	root    *NodePointer
}

func NewTreeReader(version uint32, root *NodePointer) TreeReader {
	return TreeReader{version: version, root: root}
}

func (t TreeReader) HasErr(key []byte) (bool, error) {
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return false, err
	}
	if root == nil {
		return false, nil
	}
	has, _, err := root.Has(key)
	return has, err
}

func (t TreeReader) GetErr(key []byte) ([]byte, error) {
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, nil
	}
	value, _, err := root.Get(key)
	if err != nil {
		return nil, err
	}
	return value.SafeCopy(), nil
}

func (t TreeReader) Size() int64 {
	if t.root == nil {
		return 0
	}
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return 0
	}
	return root.Size()
}

func (t TreeReader) Iterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, true, t.root)
}

func (t TreeReader) ReverseIterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, false, t.root)
}

func (t TreeReader) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (t TreeReader) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewStore(t)
}

func (t TreeReader) CacheWrapWithTrace(io.Writer, storetypes.TraceContext) storetypes.CacheWrap {
	logger.Warn("CacheWrapWithTrace called on KVStoreWrapper: tracing not implemented")
	return cachekv.NewStore(t)
}

func (t TreeReader) Get(key []byte) []byte {
	value, err := t.GetErr(key)
	if err != nil {
		panic(err)
	}
	return value
}

func (t TreeReader) Has(key []byte) bool {
	found, err := t.HasErr(key)
	if err != nil {
		panic(err)
	}
	return found
}

func (t TreeReader) Set([]byte, []byte) {
	panic("readonly store: cannot set value")
}

func (t TreeReader) Delete([]byte) {
	panic("readonly store: cannot delete")
}

func (t TreeReader) Version() uint32 {
	return t.version
}

// TODO add proof logic here

var _ storetypes.KVStore = TreeReader{}
