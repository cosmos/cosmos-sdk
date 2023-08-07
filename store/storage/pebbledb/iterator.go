package pebbledb

import (
	"bytes"

	"cosmossdk.io/store/v2"
	"github.com/cockroachdb/pebble"
)

var _ store.Iterator = (*iterator)(nil)

type iterator struct {
	source             *pebble.Iterator
	prefix, start, end []byte
	reverse            bool
	invalid            bool
}

func newPebbleDBIterator(src *pebble.Iterator, prefix, start, end []byte, reverse bool) *iterator {
	if reverse {
		if end == nil {
			_ = src.Last()
		} else {
			_ = src.SeekGE(end)

			if src.Valid() {
				eoaKey := src.Key() // end or after key
				if bytes.Compare(end, eoaKey) <= 0 {
					src.Prev()
				}

			} else {
				_ = src.Last()
			}
		}
	} else {
		if start == nil {
			_ = src.First()
		} else {
			_ = src.SeekGE(start)
		}
	}

	return &iterator{
		source:  src,
		prefix:  prefix,
		start:   start,
		end:     end,
		reverse: reverse,
		invalid: !src.Valid(),
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
	if err := itr.source.Error(); err != nil {
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
	key := itr.source.Key()

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
	key := itr.source.Key()

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
	return itr.source.Error()
}

func (itr *iterator) Close() {
	_ = itr.source.Close()
	itr.source = nil
}

func (itr *iterator) assertIsValid() {
	if !itr.Valid() {
		panic("iterator is invalid")
	}
}
