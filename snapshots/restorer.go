package snapshots

import (
	"errors"
	"fmt"
	"io"
	"time"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// Restorer is a helper that manages an asynchronous snapshot restoration process
type Restorer struct {
	height    uint64
	format    uint32
	chunks    uint32
	nextChunk uint32
	chChunks  chan<- io.ReadCloser
	chDone    <-chan error
}

// NewRestorer starts a snapshot restoration for the given target. The caller must call Close(),
// and also Complete() when the final chunks has been given.
func NewRestorer(target store.Snapshotter, height uint64, format uint32, chunks uint32) (*Restorer, error) {
	chChunks := make(chan io.ReadCloser, 4)
	chDone := make(chan error, 1)
	go func() {
		chDone <- target.Restore(height, format, chChunks)
		close(chDone)
	}()

	// Check for any initial errors from the restore. This is a bit of a code smell.
	select {
	case err := <-chDone:
		close(chChunks)
		if err == nil {
			err = errors.New("restore ended unexpectedly")
		}
		return nil, err
	case <-time.After(10 * time.Millisecond):
		return &Restorer{
			height:    height,
			format:    format,
			chunks:    chunks,
			nextChunk: 1,
			chChunks:  chChunks,
			chDone:    chDone,
		}, nil
	}
}

// Add adds a chunk to be restored. It will finalize the import when the final chunk is given,
// returning true. The returned error may not be caused by the given chunk, since the
// restore is asynchronous and since data records may span multiple chunks.
func (r *Restorer) Add(chunk io.ReadCloser) (bool, error) {
	if r.chChunks == nil {
		return false, errors.New("no restore in progress")
	}

	// check if any errors have occured so far
	select {
	case err := <-r.chDone:
		r.Close()
		if err == nil {
			err = errors.New("restore ended unexpectedly")
		}
		return false, err
	default:
	}

	// pass the chunk, and wait for completion if it was the final one
	r.chChunks <- chunk
	r.nextChunk++
	if r.nextChunk > r.chunks {
		r.Close()
		return true, <-r.chDone
	}
	return false, nil
}

// Close closes the restore, aborting it if not completed.
func (r *Restorer) Close() {
	if r != nil && r.chChunks != nil {
		close(r.chChunks)
		r.chChunks = nil
	}
}

// Expects checks if a chunk is the next expected one
func (r *Restorer) Expects(height uint64, format uint32, chunk uint32) error {
	if r == nil || r.chChunks == nil {
		return errors.New("no restore in progress")
	}
	if height != r.height {
		return fmt.Errorf("unexpected height %v, expected %v", height, r.height)
	}
	if format != r.format {
		return fmt.Errorf("unexpected format %v, expected %v", format, r.format)
	}
	if chunk != r.nextChunk {
		return fmt.Errorf("unexpected chunk %v, expected %v", chunk, r.nextChunk)
	}
	return nil
}
