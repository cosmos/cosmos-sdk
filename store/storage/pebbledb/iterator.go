package pebbledb

import (
	"bytes"
	"fmt"

	"github.com/cockroachdb/pebble"

	"cosmossdk.io/store/v2"
)

var _ store.Iterator = (*iterator)(nil)

// iterator implements the store.Iterator interface. It wraps a PebbleDB iterator
// with added MVCC key handling logic. The iterator will iterate over the key space
// in the provided domain for a given version. If a key has been written at the
// provided version, that key/value pair will be iterated over. Otherwise, the
// latest version for that key/value pair will be iterated over s.t. it's less
// than the provided version. Note:
//
// - The start key must not be empty.
// - Currently, reverse iteration is NOT supported.
type iterator struct {
	source             *pebble.Iterator
	prefix, start, end []byte
	version            uint64
	valid              bool
}

func newPebbleDBIterator(src *pebble.Iterator, prefix, mvccStart, mvccEnd []byte, version uint64) *iterator {
	// move the underlying PebbleDB iterator to the first key
	_ = src.First()

	return &iterator{
		source:  src,
		prefix:  prefix,
		start:   mvccStart,
		end:     mvccEnd,
		version: version,
		valid:   src.Valid(),
	}
}

// Domain returns the domain of the iterator. The caller must not modify the
// return values.
func (itr *iterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr *iterator) Key() []byte {
	itr.assertIsValid()

	key, _, ok := SplitMVCCKey(itr.source.Key())
	if !ok {
		// XXX: This should not happen as that would indicate we have a malformed
		// MVCC key.
		panic(fmt.Sprintf("invalid PebbleDB MVCC key: %s", itr.source.Key()))
	}

	keyCopy := make([]byte, len(key))
	_ = copy(keyCopy, key)

	return keyCopy[len(itr.prefix):]
}

func (itr *iterator) Value() []byte {
	itr.assertIsValid()
	val := itr.source.Value()

	valCopy := make([]byte, len(val))
	_ = copy(valCopy, val)

	return valCopy
}

func (itr *iterator) Next() bool {
	// First move the iterator to the next prefix, which may not correspond to the
	// desired version for that key, e.g. if the key was written at a later version,
	// so we seek back to the latest desired version, s.t. the version is <= itr.version.
	if itr.source.NextPrefix() {
		nextKey, _, ok := SplitMVCCKey(itr.source.Key())
		if !ok {
			// XXX: This should not happen as that would indicate we have a malformed
			// MVCC key.
			itr.valid = false
			return itr.valid
		}

		// Move the iterator to the closest version to the desired version, so we
		// append the current iterator key to the prefix and seek to that key.
		itr.valid = itr.source.SeekLT(MVCCEncode(nextKey, itr.version+1))
		return itr.valid
	}

	itr.valid = false
	return itr.valid
}

func (itr *iterator) Valid() bool {
	// once invalid, forever invalid
	if !itr.valid || !itr.source.Valid() {
		itr.valid = false
		return itr.valid
	}

	// if source has error, consider it invalid
	if err := itr.source.Error(); err != nil {
		itr.valid = false
		return itr.valid
	}

	// if key is at the end or past it, consider it invalid
	if end := itr.end; end != nil {
		if bytes.Compare(end, itr.Key()) <= 0 {
			itr.valid = false
			return itr.valid
		}
	}

	return true
}

func (itr *iterator) Error() error {
	return itr.source.Error()
}

func (itr *iterator) Close() {
	_ = itr.source.Close()
	itr.source = nil
	itr.valid = false
}

func (itr *iterator) assertIsValid() {
	if !itr.valid {
		panic("iterator is invalid")
	}
}

func (itr *iterator) debugRawIterate() {
	valid := itr.source.Valid()
	for valid {
		key, vBz, ok := SplitMVCCKey(itr.source.Key())
		if !ok {
			panic(fmt.Sprintf("invalid PebbleDB MVCC key: %s", itr.source.Key()))
		}

		version, err := decodeUint64Ascending(vBz)
		if err != nil {
			panic(fmt.Errorf("failed to decode version from MVCC key: %w", err))
		}

		fmt.Printf("KEY: %s, VALUE: %s, VERSION: %d\n", key, itr.source.Value(), version)

		if itr.source.NextPrefix() {
			nextKey, _, ok := SplitMVCCKey(itr.source.Key())
			if !ok {
				panic(fmt.Sprintf("invalid PebbleDB MVCC key: %s", itr.source.Key()))
			}

			valid = itr.source.SeekLT(MVCCEncode(nextKey, itr.version+1))
		} else {
			valid = false
		}
	}
}
