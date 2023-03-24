package runtime

import (
	"bytes"
	"errors"

	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
)

var _ store.KVStore = PrefixStore{}

// Store is similar with cometbft/cometbft/libs/db/prefix_db
// both gives access only to the limited subset of the store
// for convenience or safety
type PrefixStore struct {
	parent store.KVStore
	prefix []byte
}

func NewPrefixStore(parent store.KVStore, prefix []byte) PrefixStore {
	return PrefixStore{
		parent: parent,
		prefix: prefix,
	}
}

func cloneAppend(bz []byte, tail []byte) (res []byte) {
	res = make([]byte, len(bz)+len(tail))
	copy(res, bz)
	copy(res[len(bz):], tail)
	return
}

func (s PrefixStore) key(key []byte) (res []byte) {
	if key == nil {
		panic("nil key on Store")
	}
	res = cloneAppend(s.prefix, key)
	return
}

// Implements KVStore
func (s PrefixStore) Get(key []byte) ([]byte, error) {
	return s.parent.Get(s.key(key))
}

// Implements KVStore
func (s PrefixStore) Has(key []byte) (bool, error) {
	return s.parent.Has(s.key(key))
}

// Implements KVStore
func (s PrefixStore) Set(key, value []byte) error {
	// store.AssertValidKey(key)
	// store.AssertValidValue(value)
	return s.parent.Set(s.key(key), value)
}

// Implements KVStore
func (s PrefixStore) Delete(key []byte) error {
	return s.parent.Delete(s.key(key))
}

// Implements KVStore
// Check https://github.com/cometbft/cometbft/blob/master/libs/db/prefix_db.go#L106
func (s PrefixStore) Iterator(start, end []byte) (store.Iterator, error) {
	newstart := cloneAppend(s.prefix, start)

	var newend []byte
	if end == nil {
		newend = cpIncr(s.prefix)
	} else {
		newend = cloneAppend(s.prefix, end)
	}

	iter, err := s.parent.Iterator(newstart, newend)
	if err != nil {
		return nil, err
	}

	return newPrefixIterator(s.prefix, start, end, iter), nil
}

// ReverseIterator implements KVStore
// Check https://github.com/cometbft/cometbft/blob/master/libs/db/prefix_db.go#L129
func (s PrefixStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	newstart := cloneAppend(s.prefix, start)

	var newend []byte
	if end == nil {
		newend = cpIncr(s.prefix)
	} else {
		newend = cloneAppend(s.prefix, end)
	}

	iter, err := s.parent.ReverseIterator(newstart, newend)
	if err != nil {
		return nil, err
	}

	return newPrefixIterator(s.prefix, start, end, iter), nil
}

var _ store.Iterator = (*prefixIterator)(nil)

type prefixIterator struct {
	prefix []byte
	start  []byte
	end    []byte
	iter   store.Iterator
	valid  bool
}

func newPrefixIterator(prefix, start, end []byte, parent store.Iterator) *prefixIterator {
	return &prefixIterator{
		prefix: prefix,
		start:  start,
		end:    end,
		iter:   parent,
		valid:  parent.Valid() && bytes.HasPrefix(parent.Key(), prefix),
	}
}

// Implements Iterator
func (pi *prefixIterator) Domain() ([]byte, []byte) {
	return pi.start, pi.end
}

// Implements Iterator
func (pi *prefixIterator) Valid() bool {
	return pi.valid && pi.iter.Valid()
}

// Implements Iterator
func (pi *prefixIterator) Next() {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Next()")
	}

	if pi.iter.Next(); !pi.iter.Valid() || !bytes.HasPrefix(pi.iter.Key(), pi.prefix) {
		// TODO: shouldn't pi be set to nil instead?
		pi.valid = false
	}
}

// Implements Iterator
func (pi *prefixIterator) Key() (key []byte) {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Key()")
	}

	key = pi.iter.Key()
	key = stripPrefix(key, pi.prefix)

	return
}

// Implements Iterator
func (pi *prefixIterator) Value() []byte {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Value()")
	}

	return pi.iter.Value()
}

// Implements Iterator
func (pi *prefixIterator) Close() error {
	return pi.iter.Close()
}

// Error returns an error if the prefixIterator is invalid defined by the Valid
// method.
func (pi *prefixIterator) Error() error {
	if !pi.Valid() {
		return errors.New("invalid prefixIterator")
	}

	return nil
}

// copied from github.com/cometbft/cometbft/libs/db/prefix_db.go
func stripPrefix(key []byte, prefix []byte) []byte {
	if len(key) < len(prefix) || !bytes.Equal(key[:len(prefix)], prefix) {
		panic("should not happen")
	}

	return key[len(prefix):]
}

// wrapping store.PrefixEndBytes
func cpIncr(bz []byte) []byte {
	return storetypes.PrefixEndBytes(bz)
}
