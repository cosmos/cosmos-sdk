//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"bytes"

	"cosmossdk.io/store/v2"
	"github.com/linxGnu/grocksdb"
)

var _ store.Iterator = (*iterator)(nil)

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
		invalid: false,
	}
}

func (itr *iterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
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

func (itr iterator) Next() bool {
	itr.assertIsValid()

	if itr.reverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}

	return itr.Valid()
}

func (itr *iterator) Error() error {
	return itr.source.Err()
}

func (itr *iterator) Close() {
	itr.source.Close()
}

func (itr *iterator) assertIsValid() {
	if !itr.Valid() {
		panic("iterator is invalid")
	}
}
