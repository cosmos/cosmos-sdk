package internal

import (
	"io"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/iavl/internal/cachekv"
)

type TreeReader struct {
	root *NodePointer
}

func NewTreeReader(root *NodePointer) TreeReader {
	return TreeReader{root: root}
}

func (t TreeReader) HasErr(key []byte) (bool, error) {
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return false, err
	}
	// TODO optimized Has without getting the value
	value, _, err := root.Get(key)
	if err != nil {
		return false, err
	}
	return value.UnsafeBytes() != nil, nil
}

func (t TreeReader) GetErr(key []byte) ([]byte, error) {
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, err
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

var _ storetypes.KVStore = TreeReader{}
