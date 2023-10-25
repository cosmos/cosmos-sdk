package mem

import (
	"bytes"

	"github.com/tidwall/btree"
	"golang.org/x/exp/slices"

	"cosmossdk.io/store/v2"
)

var _ store.Iterator = (*iterator)(nil)

type iterator struct {
	treeItr btree.IterG[store.KVPair]
	start   []byte
	end     []byte
	reverse bool
	valid   bool
}

func newIterator(tree *btree.BTreeG[store.KVPair], start, end []byte, reverse bool) store.Iterator {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		panic(store.ErrKeyEmpty)
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		panic(store.ErrStartAfterEnd)
	}

	iter := tree.Iter()

	var valid bool
	if reverse {
		if end != nil {
			valid = iter.Seek(store.KVPair{Key: end, Value: nil})
			if !valid {
				valid = iter.Last()
			} else {
				valid = iter.Prev() // end is exclusive
			}
		} else {
			valid = iter.Last()
		}
	} else {
		if start != nil {
			valid = iter.Seek(store.KVPair{Key: start, Value: nil})
		} else {
			valid = iter.First()
		}
	}

	itr := &iterator{
		treeItr: iter,
		start:   start,
		end:     end,
		reverse: reverse,
		valid:   valid,
	}

	if itr.valid {
		itr.valid = itr.keyInRange(itr.Key())
	}

	return itr
}

// Domain returns the domain of the iterator. The caller must not modify the
// return values.
func (itr *iterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr *iterator) Valid() bool {
	return itr.valid
}

func (itr *iterator) Key() []byte {
	return slices.Clone(itr.treeItr.Item().Key)
}

func (itr *iterator) Value() []byte {
	return slices.Clone(itr.treeItr.Item().Value)
}

func (itr *iterator) Next() bool {
	if !itr.valid {
		return false
	}

	if !itr.reverse {
		itr.valid = itr.treeItr.Next()
	} else {
		itr.valid = itr.treeItr.Prev()
	}

	if itr.valid {
		itr.valid = itr.keyInRange(itr.Key())
	}

	return itr.valid
}

func (itr *iterator) Close() {
	itr.treeItr.Release()
}

func (itr *iterator) Error() error {
	return nil
}

func (itr *iterator) keyInRange(key []byte) bool {
	if !itr.reverse && itr.end != nil && bytes.Compare(key, itr.end) >= 0 {
		return false
	}
	if itr.reverse && itr.start != nil && bytes.Compare(key, itr.start) < 0 {
		return false
	}
	return true
}
