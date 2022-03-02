//go:build rocksdb

package rocksdb

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/gorocksdb"
)

type rocksDBIterator struct {
	source     *gorocksdb.Iterator
	start, end []byte
	isReverse  bool
	isInvalid  bool
	// Whether iterator has been advanced to the first element (is fully initialized)
	primed bool
}

var _ db.Iterator = (*rocksDBIterator)(nil)

func newRocksDBIterator(source *gorocksdb.Iterator, start, end []byte, isReverse bool) *rocksDBIterator {
	if isReverse {
		if end == nil {
			source.SeekToLast()
		} else {
			source.Seek(end)
			if source.Valid() {
				eoakey := moveSliceToBytes(source.Key()) // end or after key
				if bytes.Compare(end, eoakey) <= 0 {
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
	return &rocksDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
		primed:    false,
	}
}

// Domain implements Iterator.
func (itr *rocksDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Valid implements Iterator.
func (itr *rocksDBIterator) Valid() bool {
	if !itr.primed {
		return false
	}

	if itr.isInvalid {
		return false
	}

	if !itr.source.Valid() {
		itr.isInvalid = true
		return false
	}

	var (
		start = itr.start
		end   = itr.end
		key   = moveSliceToBytes(itr.source.Key())
	)
	// If key is end or past it, invalid.
	if itr.isReverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if end != nil && bytes.Compare(key, end) >= 0 {
			itr.isInvalid = true
			return false
		}
	}
	return true
}

// Key implements Iterator.
func (itr *rocksDBIterator) Key() []byte {
	itr.assertIsValid()
	return moveSliceToBytes(itr.source.Key())
}

// Value implements Iterator.
func (itr *rocksDBIterator) Value() []byte {
	itr.assertIsValid()
	return moveSliceToBytes(itr.source.Value())
}

// Next implements Iterator.
func (itr *rocksDBIterator) Next() bool {
	if !itr.primed {
		itr.primed = true
	} else {
		if itr.isReverse {
			itr.source.Prev()
		} else {
			itr.source.Next()
		}
	}
	return itr.Valid()
}

// Error implements Iterator.
func (itr *rocksDBIterator) Error() error {
	return itr.source.Err()
}

// Close implements Iterator.
func (itr *rocksDBIterator) Close() error {
	itr.source.Close()
	return nil
}

func (itr *rocksDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("iterator is invalid")
	}
}

// moveSliceToBytes will free the slice and copy out a go []byte
// This function can be applied on *Slice returned from Key() and Value()
// of an Iterator, because they are marked as freed.
func moveSliceToBytes(s *gorocksdb.Slice) []byte {
	defer s.Free()
	if !s.Exists() {
		return nil
	}
	v := make([]byte, s.Size())
	copy(v, s.Data())
	return v
}
