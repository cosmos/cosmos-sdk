//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"bytes"

	"github.com/linxGnu/grocksdb"

	corestore "cosmossdk.io/core/store"
)

var _ corestore.Iterator = (*iterator)(nil)

type iterator struct {
	source             *grocksdb.Iterator
	prefix, start, end []byte
	reverse            bool
	invalid            bool
}

func newRocksDBIterator(source *grocksdb.Iterator, prefix, start, end []byte, reverse bool) *iterator {
	if reverse {
		if end == nil {
			source.SeekToLast()
		} else {
			source.Seek(end)

			if source.Valid() {
				eoaKey := readOnlySlice(source.Key()) // end or after key
				if bytes.Compare(end, eoaKey) <= 0 {
					source.Prev()
				}
			} else {
				source.SeekToLast()
			}
		}
	} else {
		if start == nil {
			source.SeekToFirst()
		} else {
			source.Seek(start)
		}
	}

	return &iterator{
		source:  source,
		prefix:  prefix,
		start:   start,
		end:     end,
		reverse: reverse,
		invalid: !source.Valid(),
	}
}

// Domain returns the domain of the iterator. The caller must not modify the
// return values.
func (itr *iterator) Domain() ([]byte, []byte) {
	start := itr.start
	if start != nil {
		start = start[len(itr.prefix):]
		if len(start) == 0 {
			start = nil
		}
	}

	end := itr.end
	if end != nil {
		end = end[len(itr.prefix):]
		if len(end) == 0 {
			end = nil
		}
	}

	return start, end
}

func (itr *iterator) Valid() bool {
	// once invalid, forever invalid
	if itr.invalid {
		return false
	}

	// if source has error, consider it invalid
	if err := itr.source.Err(); err != nil {
		itr.invalid = true
		return false
	}

	// if source is invalid, consider it invalid
	if !itr.source.Valid() {
		itr.invalid = true
		return false
	}

	// if key is at the end or past it, consider it invalid
	start := itr.start
	end := itr.end
	key := readOnlySlice(itr.source.Key())

	if itr.reverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.invalid = true
			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.invalid = true
			return false
		}
	}

	return true
}

func (itr *iterator) Key() []byte {
	itr.assertIsValid()
	return copyAndFreeSlice(itr.source.Key())[len(itr.prefix):]
}

func (itr *iterator) Value() []byte {
	itr.assertIsValid()
	return copyAndFreeSlice(itr.source.Value())
}

func (itr *iterator) Timestamp() []byte {
	return itr.source.Timestamp().Data()
}

func (itr iterator) Next() {
	if itr.invalid {
		return
	}

	if itr.reverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

func (itr *iterator) Error() error {
	return itr.source.Err()
}

func (itr *iterator) Close() error {
	itr.source.Close()
	itr.source = nil
	itr.invalid = true

	return nil
}

func (itr *iterator) assertIsValid() {
	if itr.invalid {
		panic("iterator is invalid")
	}
}
