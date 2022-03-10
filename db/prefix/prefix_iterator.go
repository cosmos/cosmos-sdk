package prefix

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/db"
)

// IteratePrefix is a convenience function for iterating over a key domain
// restricted by prefix.
func IteratePrefix(dbr db.DBReader, prefix []byte) (db.Iterator, error) {
	var start, end []byte
	if len(prefix) != 0 {
		start = prefix
		end = cpIncr(prefix)
	}
	itr, err := dbr.Iterator(start, end)
	if err != nil {
		return nil, err
	}
	return itr, nil
}

// Strips prefix while iterating from Iterator.
type prefixDBIterator struct {
	prefix []byte
	start  []byte
	end    []byte
	source db.Iterator
	err    error
}

var _ db.Iterator = (*prefixDBIterator)(nil)

func newPrefixIterator(prefix, start, end []byte, source db.Iterator) *prefixDBIterator {
	return &prefixDBIterator{
		prefix: prefix,
		start:  start,
		end:    end,
		source: source,
	}
}

// Domain implements Iterator.
func (itr *prefixDBIterator) Domain() (start, end []byte) {
	return itr.start, itr.end
}

func (itr *prefixDBIterator) valid() bool {
	if itr.err != nil {
		return false
	}

	key := itr.source.Key()
	if len(key) < len(itr.prefix) || !bytes.Equal(key[:len(itr.prefix)], itr.prefix) {
		itr.err = fmt.Errorf("received invalid key from backend: %x (expected prefix %x)",
			key, itr.prefix)
		return false
	}

	return true
}

// Next implements Iterator.
func (itr *prefixDBIterator) Next() bool {
	if !itr.source.Next() {
		return false
	}
	key := itr.source.Key()
	if !bytes.HasPrefix(key, itr.prefix) {
		return false
	}
	// Empty keys are not allowed, so if a key exists in the database that exactly matches the
	// prefix we need to skip it.
	if bytes.Equal(key, itr.prefix) {
		return itr.Next()
	}
	return true
}

// Next implements Iterator.
func (itr *prefixDBIterator) Key() []byte {
	itr.assertIsValid()
	key := itr.source.Key()
	return key[len(itr.prefix):] // we have checked the key in Valid()
}

// Value implements Iterator.
func (itr *prefixDBIterator) Value() []byte {
	itr.assertIsValid()
	return itr.source.Value()
}

// Error implements Iterator.
func (itr *prefixDBIterator) Error() error {
	if err := itr.source.Error(); err != nil {
		return err
	}
	return itr.err
}

// Close implements Iterator.
func (itr *prefixDBIterator) Close() error {
	return itr.source.Close()
}

func (itr *prefixDBIterator) assertIsValid() {
	if !itr.valid() {
		panic("iterator is invalid")
	}
}
