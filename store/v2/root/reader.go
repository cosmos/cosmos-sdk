package root

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

var (
	_ corestore.Reader    = (*Reader)(nil)
	_ corestore.ReaderMap = (*ReaderMap)(nil)
)

// ReaderMap defines an adapter around a RootStore that only exposes read-only
// operations. This is useful for exposing a read-only view of the RootStore at
// a specific version in history, which could also be the latest state.
type ReaderMap struct {
	vReader store.VersionedReader
	version uint64
}

func NewReaderMap(v uint64, vr store.VersionedReader) *ReaderMap {
	return &ReaderMap{
		vReader: vr,
		version: v,
	}
}

func (rm *ReaderMap) GetReader(actor []byte) (corestore.Reader, error) {
	return NewReader(rm.version, rm.vReader, actor), nil
}

// Reader represents a read-only adapter for accessing data from the root store.
type Reader struct {
	version uint64                // The version of the data.
	vReader store.VersionedReader // The root store to read data from.
	actor   []byte                // The actor associated with the data.
}

func NewReader(v uint64, vr store.VersionedReader, actor []byte) *Reader {
	return &Reader{
		version: v,
		vReader: vr,
		actor:   actor,
	}
}

func (roa *Reader) Has(key []byte) (bool, error) {
	val, err := roa.vReader.Has(roa.actor, roa.version, key)
	if err != nil {
		return false, err
	}

	return val, nil
}

func (roa *Reader) Get(key []byte) ([]byte, error) {
	return roa.vReader.Get(roa.actor, roa.version, key)
}

func (roa *Reader) Iterator(start, end []byte) (corestore.Iterator, error) {
	return roa.vReader.Iterator(roa.actor, roa.version, start, end)
}

func (roa *Reader) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return roa.vReader.ReverseIterator(roa.actor, roa.version, start, end)
}
