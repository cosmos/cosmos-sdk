// Prefixed DB reader/writer types let you namespace multiple DBs within a single DB.

package prefix

import (
	"github.com/cosmos/cosmos-sdk/db"
)

// prefixed Reader
type prefixR struct {
	db     db.DBReader
	prefix []byte
}

// prefixed ReadWriter
type prefixRW struct {
	db     db.DBReadWriter
	prefix []byte
}

// prefixed Writer
type prefixW struct {
	db     db.DBWriter
	prefix []byte
}

var (
	_ db.DBReader     = (*prefixR)(nil)
	_ db.DBReadWriter = (*prefixRW)(nil)
	_ db.DBWriter     = (*prefixW)(nil)
)

// NewPrefixReader returns a DBReader that only has access to the subset of DB keys
// that contain the given prefix.
func NewPrefixReader(dbr db.DBReader, prefix []byte) prefixR {
	return prefixR{
		prefix: prefix,
		db:     dbr,
	}
}

// NewPrefixReadWriter returns a DBReader that only has access to the subset of DB keys
// that contain the given prefix.
func NewPrefixReadWriter(dbrw db.DBReadWriter, prefix []byte) prefixRW {
	return prefixRW{
		prefix: prefix,
		db:     dbrw,
	}
}

// NewPrefixWriter returns a DBWriter that reads/writes only from the subset of DB keys
// that contain the given prefix
func NewPrefixWriter(dbw db.DBWriter, prefix []byte) prefixW {
	return prefixW{
		prefix: prefix,
		db:     dbw,
	}
}

func prefixed(prefix, key []byte) []byte {
	return append(cp(prefix), key...)
}

// Get implements DBReader.
func (pdb prefixR) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, db.ErrKeyEmpty
	}
	return pdb.db.Get(prefixed(pdb.prefix, key))
}

// Has implements DBReader.
func (pdb prefixR) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, db.ErrKeyEmpty
	}
	return pdb.db.Has(prefixed(pdb.prefix, key))
}

// Iterator implements DBReader.
func (pdb prefixR) Iterator(start, end []byte) (db.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, db.ErrKeyEmpty
	}

	var pend []byte
	if end == nil {
		pend = cpIncr(pdb.prefix)
	} else {
		pend = prefixed(pdb.prefix, end)
	}
	itr, err := pdb.db.Iterator(prefixed(pdb.prefix, start), pend)
	if err != nil {
		return nil, err
	}
	return newPrefixIterator(pdb.prefix, start, end, itr), nil
}

// ReverseIterator implements DBReader.
func (pdb prefixR) ReverseIterator(start, end []byte) (db.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, db.ErrKeyEmpty
	}

	var pend []byte
	if end == nil {
		pend = cpIncr(pdb.prefix)
	} else {
		pend = prefixed(pdb.prefix, end)
	}
	ritr, err := pdb.db.ReverseIterator(prefixed(pdb.prefix, start), pend)
	if err != nil {
		return nil, err
	}
	return newPrefixIterator(pdb.prefix, start, end, ritr), nil
}

// Discard implements DBReader.
func (pdb prefixR) Discard() error { return pdb.db.Discard() }

// Set implements DBReadWriter.
func (pdb prefixRW) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return pdb.db.Set(prefixed(pdb.prefix, key), value)
}

// Delete implements DBReadWriter.
func (pdb prefixRW) Delete(key []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return pdb.db.Delete(prefixed(pdb.prefix, key))
}

// Get implements DBReadWriter.
func (pdb prefixRW) Get(key []byte) ([]byte, error) {
	return NewPrefixReader(pdb.db, pdb.prefix).Get(key)
}

// Has implements DBReadWriter.
func (pdb prefixRW) Has(key []byte) (bool, error) {
	return NewPrefixReader(pdb.db, pdb.prefix).Has(key)
}

// Iterator implements DBReadWriter.
func (pdb prefixRW) Iterator(start, end []byte) (db.Iterator, error) {
	return NewPrefixReader(pdb.db, pdb.prefix).Iterator(start, end)
}

// ReverseIterator implements DBReadWriter.
func (pdb prefixRW) ReverseIterator(start, end []byte) (db.Iterator, error) {
	return NewPrefixReader(pdb.db, pdb.prefix).ReverseIterator(start, end)
}

// Close implements DBReadWriter.
func (pdb prefixRW) Commit() error { return pdb.db.Commit() }

// Discard implements DBReadWriter.
func (pdb prefixRW) Discard() error { return pdb.db.Discard() }

// Set implements DBReadWriter.
func (pdb prefixW) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return pdb.db.Set(prefixed(pdb.prefix, key), value)
}

// Delete implements DBWriter.
func (pdb prefixW) Delete(key []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return pdb.db.Delete(prefixed(pdb.prefix, key))
}

// Close implements DBWriter.
func (pdb prefixW) Commit() error { return pdb.db.Commit() }

// Discard implements DBReadWriter.
func (pdb prefixW) Discard() error { return pdb.db.Discard() }

func cp(bz []byte) (ret []byte) {
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}

// Returns a new slice of the same length (big endian), but incremented by one.
// Returns nil on overflow (e.g. if bz bytes are all 0xFF)
// CONTRACT: len(bz) > 0
func cpIncr(bz []byte) (ret []byte) {
	if len(bz) == 0 {
		panic("cpIncr expects non-zero bz length")
	}
	ret = cp(bz)
	for i := len(bz) - 1; i >= 0; i-- {
		if ret[i] < byte(0xFF) {
			ret[i]++
			return
		}
		ret[i] = byte(0x00)
		if i == 0 {
			// Overflow
			return nil
		}
	}
	return nil
}
