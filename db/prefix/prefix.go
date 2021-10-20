package prefix

import (
	dbm "github.com/cosmos/cosmos-sdk/db"
)

// Prefix Reader/Writer lets you namespace multiple DBs within a single DB.
type prefixR struct {
	db     dbm.DBReader
	prefix []byte
}

type prefixRW struct {
	db     dbm.DBReadWriter
	prefix []byte
}

var _ dbm.DBReader = (*prefixR)(nil)
var _ dbm.DBReadWriter = (*prefixRW)(nil)

func NewPrefixReader(db dbm.DBReader, prefix []byte) prefixR {
	return prefixR{
		prefix: prefix,
		db:     db,
	}
}

func NewPrefixReadWriter(db dbm.DBReadWriter, prefix []byte) prefixRW {
	return prefixRW{
		prefix: prefix,
		db:     db,
	}
}

func prefixed(prefix, key []byte) []byte {
	return append(prefix, key...)
}

// Get implements DBReader.
func (pdb prefixR) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, dbm.ErrKeyEmpty
	}
	return pdb.db.Get(prefixed(pdb.prefix, key))
}

// Has implements DBReader.
func (pdb prefixR) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, dbm.ErrKeyEmpty
	}
	return pdb.db.Has(prefixed(pdb.prefix, key))
}

// Iterator implements DBReader.
func (pdb prefixR) Iterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
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
func (pdb prefixR) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
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
		return dbm.ErrKeyEmpty
	}
	return pdb.db.Set(prefixed(pdb.prefix, key), value)
}

// Delete implements DBReadWriter.
func (pdb prefixRW) Delete(key []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
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
func (pdb prefixRW) Iterator(start, end []byte) (dbm.Iterator, error) {
	return NewPrefixReader(pdb.db, pdb.prefix).Iterator(start, end)
}

// ReverseIterator implements DBReadWriter.
func (pdb prefixRW) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	return NewPrefixReader(pdb.db, pdb.prefix).ReverseIterator(start, end)
}

// Close implements DBReadWriter.
func (pdb prefixRW) Commit() error { return pdb.db.Commit() }

// Discard implements DBReadWriter.
func (pdb prefixRW) Discard() error { return pdb.db.Discard() }

// Returns a slice of the same length (big endian), but incremented by one.
// Returns nil on overflow (e.g. if bz bytes are all 0xFF)
// CONTRACT: len(bz) > 0
func cpIncr(bz []byte) (ret []byte) {
	if len(bz) == 0 {
		panic("cpIncr expects non-zero bz length")
	}
	ret = make([]byte, len(bz))
	copy(ret, bz)
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
