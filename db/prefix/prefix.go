// Prefixed DB reader/writer types let you namespace multiple DBs within a single DB.

package prefix

import (
	"github.com/cosmos/cosmos-sdk/db"
)

// prefixed Reader
type Reader struct {
	db     db.Reader
	prefix []byte
}

// prefixed ReadWriter
type ReadWriter struct {
	db     db.ReadWriter
	prefix []byte
}

// prefixed Writer
type Writer struct {
	db     db.Writer
	prefix []byte
}

var (
	_ db.Reader     = (*Reader)(nil)
	_ db.ReadWriter = (*ReadWriter)(nil)
	_ db.Writer     = (*Writer)(nil)
)

// NewReadereader returns a DBReader that only has access to the subset of DB keys
// that contain the given prefix.
func NewReader(dbr db.Reader, prefix []byte) Reader {
	return Reader{
		prefix: prefix,
		db:     dbr,
	}
}

// NewReadWriter returns a DBReader that only has access to the subset of DB keys
// that contain the given prefix.
func NewReadWriter(dbrw db.ReadWriter, prefix []byte) ReadWriter {
	return ReadWriter{
		prefix: prefix,
		db:     dbrw,
	}
}

// NewWriterriter returns a DBWriter that reads/writes only from the subset of DB keys
// that contain the given prefix
func NewWriter(dbw db.Writer, prefix []byte) Writer {
	return Writer{
		prefix: prefix,
		db:     dbw,
	}
}

func prefixed(prefix, key []byte) []byte {
	return append(cp(prefix), key...)
}

// Get implements DBReader.
func (pdb Reader) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, db.ErrKeyEmpty
	}
	return pdb.db.Get(prefixed(pdb.prefix, key))
}

// Has implements DBReader.
func (pdb Reader) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, db.ErrKeyEmpty
	}
	return pdb.db.Has(prefixed(pdb.prefix, key))
}

// Iterator implements DBReader.
func (pdb Reader) Iterator(start, end []byte) (db.Iterator, error) {
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
func (pdb Reader) ReverseIterator(start, end []byte) (db.Iterator, error) {
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
func (pdb Reader) Discard() error { return pdb.db.Discard() }

// Set implements DBReadWriter.
func (pdb ReadWriter) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return pdb.db.Set(prefixed(pdb.prefix, key), value)
}

// Delete implements DBReadWriter.
func (pdb ReadWriter) Delete(key []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return pdb.db.Delete(prefixed(pdb.prefix, key))
}

// Get implements DBReadWriter.
func (pdb ReadWriter) Get(key []byte) ([]byte, error) {
	return NewReader(pdb.db, pdb.prefix).Get(key)
}

// Has implements DBReadWriter.
func (pdb ReadWriter) Has(key []byte) (bool, error) {
	return NewReader(pdb.db, pdb.prefix).Has(key)
}

// Iterator implements DBReadWriter.
func (pdb ReadWriter) Iterator(start, end []byte) (db.Iterator, error) {
	return NewReader(pdb.db, pdb.prefix).Iterator(start, end)
}

// ReverseIterator implements DBReadWriter.
func (pdb ReadWriter) ReverseIterator(start, end []byte) (db.Iterator, error) {
	return NewReader(pdb.db, pdb.prefix).ReverseIterator(start, end)
}

// Close implements DBReadWriter.
func (pdb ReadWriter) Commit() error { return pdb.db.Commit() }

// Discard implements DBReadWriter.
func (pdb ReadWriter) Discard() error { return pdb.db.Discard() }

// Set implements DBReadWriter.
func (pdb Writer) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return pdb.db.Set(prefixed(pdb.prefix, key), value)
}

// Delete implements DBWriter.
func (pdb Writer) Delete(key []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return pdb.db.Delete(prefixed(pdb.prefix, key))
}

// Close implements DBWriter.
func (pdb Writer) Commit() error { return pdb.db.Commit() }

// Discard implements DBReadWriter.
func (pdb Writer) Discard() error { return pdb.db.Discard() }

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
