package snapshots

import (
	"errors"
	"io"
	"time"

	"github.com/cosmos/cosmos-sdk/snapshots/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

// Restorer is a helper that manages an asynchronous snapshot restoration process.
type Restorer struct {
	snapshot   types.Snapshot
	chChunks   chan<- io.ReadCloser
	chDone     <-chan error
	chunksSeen int
}

// NewRestorer starts a snapshot restoration for the given target. The caller must call Close(),
// and also Complete() when the final chunks has been given.
func NewRestorer(snapshot types.Snapshot, target store.Snapshotter) (*Restorer, error) {
	chChunks := make(chan io.ReadCloser, 4)
	chDone := make(chan error, 1)
	go func() {
		chDone <- target.Restore(snapshot.Height, snapshot.Format, chChunks)
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
			snapshot: snapshot,
			chChunks: chChunks,
			chDone:   chDone,
		}, nil
	}
}

// Add adds a chunk to be restored. It will finalize the import when the final chunk is given,
// returning true. The returned error may not be caused by the given chunk, since the
// restore is asynchronous and since data records may span multiple chunks.
// FIXME This should probably verify chunk checksums as well.
func (r *Restorer) Add(chunk io.ReadCloser) (bool, error) {
	if r == nil || r.chChunks == nil {
		return false, errors.New("no restore in progress")
	}

	// check if any errors have occurred so far
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
	r.chunksSeen++
	if r.chunksSeen >= len(r.snapshot.Chunks) {
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

// Snapshot returns the snapshot being restored
func (r *Restorer) Snapshot() *types.Snapshot {
	if r == nil {
		return nil
	}
	return &r.snapshot
}
