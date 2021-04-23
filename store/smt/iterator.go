package smt

import (
	"bytes"

	dbm "github.com/tendermint/tm-db"
)

type Iterator struct {
	store *Store
	iter  dbm.Iterator
}

func indexKey(key []byte) []byte {
	return append(indexPrefix, key...)
}

func plainKey(key []byte) []byte {
	return key[prefixLen:]
}

func startKey(key []byte) []byte {
	if key == nil {
		return dataPrefix
	}
	return dataKey(key)
}

func endKey(key []byte) []byte {
	if key == nil {
		return indexPrefix
	}
	return dataKey(key)
}

func newIterator(s *Store, start, end []byte, reverse bool) (*Iterator, error) {
	start = startKey(start)
	end = endKey(end)
	var i dbm.Iterator
	var err error
	if reverse {
		i, err = s.db.ReverseIterator(start, end)
	} else {
		i, err = s.db.Iterator(start, end)
	}
	if err != nil {
		return nil, err
	}
	return &Iterator{store: s, iter: i}, nil
}

// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
// CONTRACT: start, end readonly []byte
func (i *Iterator) Domain() (start []byte, end []byte) {
	start, end = i.iter.Domain()
	if bytes.Equal(start, dataPrefix) {
		start = nil
	} else {
		start = plainKey(start)
	}
	if bytes.Equal(end, indexPrefix) {
		end = nil
	} else {
		end = plainKey(end)
	}
	return start, end
}

// Valid returns whether the current iterator is valid. Once invalid, the Iterator remains
// invalid forever.
func (i *Iterator) Valid() bool {
	return i.iter.Valid()
}

// Next moves the iterator to the next key in the database, as defined by order of iteration.
// If Valid returns false, this method will panic.
func (i *Iterator) Next() {
	i.iter.Next()
}

// Key returns the key at the current position. Panics if the iterator is invalid.
// CONTRACT: key readonly []byte
func (i *Iterator) Key() (key []byte) {
	return plainKey(i.iter.Key())
}

// Value returns the value at the current position. Panics if the iterator is invalid.
// CONTRACT: value readonly []byte
func (i *Iterator) Value() (value []byte) {
	return i.store.Get(i.Key())
}

// Error returns the last error encountered by the iterator, if any.
func (i *Iterator) Error() error {
	return i.iter.Error()
}

// Close closes the iterator, relasing any allocated resources.
func (i *Iterator) Close() error {
	return i.iter.Close()
}
