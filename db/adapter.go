package db

import "errors"

type readerRWAdapter struct{ DBReader }

var (
	// ErrReadOnly is returned when a write operation is attempted on a read-only transaction.
	ErrReadOnly = errors.New("Cannot modify read-only transaction")
)

// Returns a ReadWriter that forwards to a reader and errors if writes are
// attempted. Can be used to pass a Reader when a ReadWriter is expected
// but no writes will actually occur.
func NewReadWriterFromReader(r DBReader) DBReadWriter {
	return readerRWAdapter{r}
}

func (readerRWAdapter) Set([]byte, []byte) error {
	return ErrReadOnly
}

func (readerRWAdapter) Delete([]byte) error {
	return ErrReadOnly
}

func (rw readerRWAdapter) Commit() error {
	rw.Discard()
	return nil
}
