package db

type readerRWAdapter struct{ Reader }

// ReaderAsReadWriter returns a ReadWriter that forwards to a reader and errors if writes are
// attempted. Can be used to pass a Reader when a ReadWriter is expected
// but no writes will actually occur.
func ReaderAsReadWriter(r Reader) ReadWriter {
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
