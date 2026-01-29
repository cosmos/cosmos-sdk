package internal

import (
	storetypes "cosmossdk.io/store/types"
)

type TreeReader struct {
	root *NodePointer
}

func (t TreeReader) Has(key []byte) (bool, error) {
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

func (t TreeReader) Get(key []byte) ([]byte, error) {
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
