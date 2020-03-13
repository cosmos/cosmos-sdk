package rootmulti

import (
	"errors"
	"io"
)

// chunkWriter reads an input stream, splits it into fixed-size chunks, and writes them to a
// sequence of io.ReadClosers via a channel
type chunkWriter struct {
	ch        chan<- io.ReadCloser
	pipe      *io.PipeWriter
	chunkSize uint64
	written   uint64
	closed    bool
}

// newChunkWriter creates a new chunkWriter
func newChunkWriter(ch chan<- io.ReadCloser, chunkSize uint64) (*chunkWriter, error) {
	if chunkSize == 0 {
		return nil, errors.New("chunk size cannot be 0")
	}
	return &chunkWriter{
		ch:        ch,
		chunkSize: chunkSize,
	}, nil
}

// chunk creates a new chunk
func (w *chunkWriter) chunk() error {
	if w.pipe != nil {
		err := w.pipe.Close()
		if err != nil {
			return err
		}
	}
	pr, pw := io.Pipe()
	w.ch <- pr
	w.pipe = pw
	w.written = 0
	return nil
}

// Close implements io.Closer
func (w *chunkWriter) Close() error {
	if !w.closed {
		w.closed = true
		close(w.ch)
		return w.pipe.Close()
	}
	return nil
}

// CloseWithError closes the writer and sends an error to the reader
func (w *chunkWriter) CloseWithError(err error) {
	if !w.closed {
		w.closed = true
		close(w.ch)
		w.pipe.CloseWithError(err)
	}
}

// Write implements io.Writer
func (w *chunkWriter) Write(data []byte) (int, error) {
	if w.closed {
		return 0, errors.New("cannot write to closed chunkWriter")
	}
	nTotal := 0
	for len(data) > 0 {
		if w.pipe == nil || w.written >= w.chunkSize {
			err := w.chunk()
			if err != nil {
				return nTotal, err
			}
		}
		writeSize := w.chunkSize - w.written
		if writeSize > uint64(len(data)) {
			writeSize = uint64(len(data))
		}
		n, err := w.pipe.Write(data[:writeSize])
		w.written += uint64(n)
		nTotal += n
		if err != nil {
			return nTotal, err
		}
		data = data[writeSize:]
	}
	return nTotal, nil
}

// chunkReader reads chunks from a channel of io.ReadClosers and outputs them as an io.Reader
type chunkReader struct {
	ch     <-chan io.ReadCloser
	reader io.ReadCloser
}

// newChunkReader creates a new chunkReader
func newChunkReader(ch <-chan io.ReadCloser) *chunkReader {
	return &chunkReader{ch: ch}
}

// next fetches the next chunk from the channel, or returns io.EOF if there are no more chunks
func (r *chunkReader) next() error {
	reader, ok := <-r.ch
	if !ok {
		return io.EOF
	}
	r.reader = reader
	return nil
}

// Close implements io.ReadCloser
func (r *chunkReader) Close() error {
	if r.reader != nil {
		err := r.reader.Close()
		r.reader = nil
		return err
	}
	return nil
}

// Read implements io.Reader
func (r *chunkReader) Read(p []byte) (int, error) {
	if r.reader == nil {
		err := r.next()
		if err != nil {
			return 0, err
		}
	}
	n, err := r.reader.Read(p)
	if err == io.EOF {
		err = r.reader.Close()
		r.reader = nil
		if err != nil {
			return 0, err
		}
		return r.Read(p)
	}
	return n, err
}
